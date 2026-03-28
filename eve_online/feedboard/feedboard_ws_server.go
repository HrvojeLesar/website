package feedboard

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"text/template"
	"time"

	zkillboard "github.com/HrvojeLesar/website/eve_online/zkillboard"
	"github.com/coder/websocket"
)

type feedboardWebsocketServer struct {
	KillmailReceiverChannel chan *zkillboard.Killmail
	subscribers             map[*feedboardSubscriber]struct{}
	subscribersMutex        sync.Mutex
	templateBuilderMutex    sync.Mutex
}

type feedboardWebsocketServerBuilder struct{}

var FeedboardWebsocketServerBuilder feedboardWebsocketServerBuilder

func (builder *feedboardWebsocketServerBuilder) New(killmailChannel chan *zkillboard.Killmail) feedboardWebsocketServer {
	return feedboardWebsocketServer{
		KillmailReceiverChannel: killmailChannel,
	}
}

func (server *feedboardWebsocketServer) SubscribeHandler(writer http.ResponseWriter, request *http.Request) {
	err := server.subscribe(request.Context(), writer, request)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		slog.Error("Websocket subscribe error:", "error", err)
		return
	}
}

func (server *feedboardWebsocketServer) StartKillmailListener(killmailchan chan *zkillboard.Killmail) {
	go func() {
		var templateBuffer bytes.Buffer
		for {
			killmail := <-killmailchan
			if killmail.Victim == nil || killmail.FinalBlow == nil {
				slog.Error("Skipping templating killmail, missing victim or final blow character", "Victim", killmail.Victim, "FinalBlow", killmail.FinalBlow)
				continue
			}
			server.templateBuilderMutex.Lock()
			slog.Debug("Received killmail", "id", killmail.KillmailID)
			templ := template.Must(template.ParseFiles("templates/feedboard_item.html"))
			err := templ.Execute(&templateBuffer, &killmail)
			if err != nil {
				slog.Error("Failed to parse template", "error", err)
				continue
			}
			server.sendTemplate(templateBuffer.Bytes())
			templateBuffer.Reset()
			server.templateBuilderMutex.Unlock()
		}
	}()
}

func (server *feedboardWebsocketServer) subscribe(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	var connectionMutex sync.Mutex
	acceptedConnection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		return err
	}
	defer acceptedConnection.CloseNow()

	closed := false
	subscriber := FeedboardSubscriberBuilder.NewSubscriber(func() {
		connectionMutex.Lock()
		defer connectionMutex.Unlock()
		closed = true
		if acceptedConnection != nil {
			acceptedConnection.Close(websocket.StatusPolicyViolation, "Connection too slow to keep up with messages")
		}
	})

	server.addSubscriber(&subscriber)
	defer server.removeSubscriber(&subscriber)

	connectionMutex.Lock()
	if closed {
		connectionMutex.Unlock()
		return net.ErrClosed
	}
	connectionMutex.Lock()

	ctx = acceptedConnection.CloseRead(ctx)

	for {
		select {
		case msg := <-subscriber.MessagesChannel:
			{
				writeContext, cancelWriteContext := context.WithTimeout(ctx, 5*time.Second)
				err := acceptedConnection.Write(writeContext, websocket.MessageText, msg)
				cancelWriteContext()
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (server *feedboardWebsocketServer) addSubscriber(subscriber *feedboardSubscriber) {
	slog.Debug("Added a new subscriber")
	server.subscribersMutex.Lock()
	server.subscribers[subscriber] = struct{}{}
	server.subscribersMutex.Unlock()
}

func (server *feedboardWebsocketServer) removeSubscriber(subscriber *feedboardSubscriber) {
	slog.Debug("Removed a subscriber")
	server.subscribersMutex.Lock()
	delete(server.subscribers, subscriber)
	server.subscribersMutex.Unlock()
}

func (server *feedboardWebsocketServer) sendTemplate(template []byte) {
	server.subscribersMutex.Lock()
	defer server.subscribersMutex.Unlock()

	for subscriber := range server.subscribers {
		select {
		case subscriber.MessagesChannel <- template:
		default:
			go subscriber.CloseSlowFunc()
		}
	}
}
