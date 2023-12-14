package main

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type ZkillWebsocketFilter struct {
	Action  string `json:"action"`
	Channel string `json:"channel"`
}

type ZkillWebsocketSimpleKillmail struct {
	Action        *string `json:"action"`
	KillId        int64   `json:"killID"`
	Hash          *string `json:"hash"`
	CharacterId   int64   `json:"character_id"`
	CorporationId int64   `json:"corporation_id"`
	AllianceId    int64   `json:"alliance_id"`
	ShipTypeId    int64   `json:"ship_type_id"`
	Url           *string `json:"url"`
}

type ZkillKmChannel struct {
	*ZkillWebsocketSimpleKillmail
	Error error
}

type Context struct {
	Context    context.Context
	CancelFunc context.CancelFunc
}

type ZkillWebsocket struct {
	Filter      *ZkillWebsocketFilter
	ReadContext Context
	DialContext Context
	connection  *websocket.Conn
}

func NewZkillWebsocket(filter ZkillWebsocketFilter) *ZkillWebsocket {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	readCtx, readCancel := context.WithCancel(context.Background())
	return &ZkillWebsocket{
		Filter:      &filter,
		DialContext: Context{Context: dialCtx, CancelFunc: dialCancel},
		ReadContext: Context{Context: readCtx, CancelFunc: readCancel},
	}
}

func (zk *ZkillWebsocket) Connect() error {
	connection, _, err := websocket.Dial(zk.DialContext.Context, "wss://zkillboard.com/websocket/", nil)
	defer zk.DialContext.CancelFunc()
	if err != nil {
		return err
	}
	zk.connection = connection

	err = wsjson.Write(zk.DialContext.Context, zk.connection, zk.Filter)
	if err != nil {
		return err
	}

	return nil
}

func (zk *ZkillWebsocket) Disconnect() error {
	err := zk.connection.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		return err
	}
	return nil
}

func (zk *ZkillWebsocket) KillmailChan() chan *ZkillKmChannel {
	var v ZkillWebsocketSimpleKillmail
	kmChan := make(chan *ZkillKmChannel)
	go func() {
		for {
			err := wsjson.Read(zk.ReadContext.Context, zk.connection, &v)
			log.Println("New killmail recieved")
			if err != nil {
				kmChan <- &ZkillKmChannel{
					ZkillWebsocketSimpleKillmail: nil,
					Error:                        err,
				}
				return
			}
			kmChan <- &ZkillKmChannel{
				ZkillWebsocketSimpleKillmail: &v,
				Error:                        err,
			}
		}
	}()

	go func() {
		<-zk.ReadContext.Context.Done()
		log.Println("Websocket connection cancelled")
		kmChan <- &ZkillKmChannel{
			ZkillWebsocketSimpleKillmail: nil,
			Error:                        errors.New("Websocket connection closed"),
		}
	}()

	return kmChan
}

type ZkillWebsocketManager struct {
	ZkillWebsocketFilter
	Backoff  backoff.BackOff
	Callback func(km *ZkillWebsocketSimpleKillmail)
}

func NewZkillWebsocketManager(killmailCallback func(km *ZkillWebsocketSimpleKillmail), filter ZkillWebsocketFilter) *ZkillWebsocketManager {
	return &ZkillWebsocketManager{
		ZkillWebsocketFilter: filter,
		Backoff:              backoff.NewExponentialBackOff(),
		Callback:             killmailCallback,
	}
}

func (zk *ZkillWebsocketManager) Run() {
	go func() {
		backoff.RetryNotify(zk.connectAndReadWebsocket, zk.Backoff,
			func(err error, d time.Duration) {
				log.Println("Zkill websocket error: ", err)
				log.Println("Trying reconnect in: ", d)
			})
	}()
}

func (zk *ZkillWebsocketManager) connectAndReadWebsocket() (err error) {
	ws := NewZkillWebsocket(zk.ZkillWebsocketFilter)
	err = ws.Connect()
	if err != nil {
		return err
	}

	zk.Backoff.Reset()

	killmailChan := ws.KillmailChan()
	for {
		killmail := <-killmailChan
		if killmail.Error != nil {
			return killmail.Error
		}

		zk.Callback(killmail.ZkillWebsocketSimpleKillmail)
	}
}

type feedboardSubscriber struct {
	msgs      chan []byte
	closeSlow func()
}

type FeedboardWebsocketServer struct {
	KillmailChan            chan []FeedboardKillmail
	subscriberMessageBuffer int
	subscribersMu           sync.Mutex
	subscribers             map[*feedboardSubscriber]struct{}
}

func newFeedboardWebsocketServer() *FeedboardWebsocketServer {
	return &FeedboardWebsocketServer{
		KillmailChan:            make(chan []FeedboardKillmail),
		subscriberMessageBuffer: 16,
		subscribers:             make(map[*feedboardSubscriber]struct{}),
	}
}

func (fws *FeedboardWebsocketServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := fws.subscribe(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		log.Println(err)
		return
	}
}

func (fws *FeedboardWebsocketServer) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	s := &feedboardSubscriber{
		msgs: make(chan []byte, fws.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "Connection too slow to kueep up with messages")
			}
		},
	}

	fws.addSubscriber(s)
	defer fws.deleteSubscriber(s)

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, 5*time.Second, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (fws *FeedboardWebsocketServer) addSubscriber(s *feedboardSubscriber) {
	log.Println("New sub")
	fws.subscribersMu.Lock()
	fws.subscribers[s] = struct{}{}
	fws.subscribersMu.Unlock()
}

func (fws *FeedboardWebsocketServer) deleteSubscriber(s *feedboardSubscriber) {
	fws.subscribersMu.Lock()
	delete(fws.subscribers, s)
	fws.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func (fws *FeedboardWebsocketServer) KillmailListener() {
	go func() {
		var templateBuffer bytes.Buffer
		for {
            log.Println("WS sub number: ", len(fws.subscribers))
			templateBuffer.Reset()
			killmails := <-fws.KillmailChan
			if len(killmails) < 1 {
				continue
			}
			templ := template.Must(template.ParseFiles("templates/feedboard_item.html"))
			err := templ.Execute(&templateBuffer, &killmails[0])
			if err != nil {
				log.Println(err)
				continue
			}
			fws.sendTemplate(templateBuffer.Bytes())
		}
	}()
}

func (fws *FeedboardWebsocketServer) sendTemplate(template []byte) {
	fws.subscribersMu.Lock()
	defer fws.subscribersMu.Unlock()

	for s := range fws.subscribers {
		select {
		case s.msgs <- template:
		default:
			go s.closeSlow()
		}
	}
}
