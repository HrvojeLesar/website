package feedboard

import (
	"testing"
	"time"

	zkillboard "github.com/HrvojeLesar/website/eve_online/zkillboard"
)

func TestNewWebsocketServer(t *testing.T) {
	t.Run("creates server with channels", func(t *testing.T) {
		killmailCh := make(chan *zkillboard.Killmail)
		cacheUpdatedCh := make(chan zkillboard.KillmailCollection)

		server := FeedboardWebsocketServerBuilder.New(killmailCh, cacheUpdatedCh)

		if server.KillmailReceiverChannel != killmailCh {
			t.Error("KillmailReceiverChannel not set")
		}
		if server.CacheUpdatedChannel != cacheUpdatedCh {
			t.Error("CacheUpdatedChannel not set")
		}
		if server.subscribers == nil {
			t.Error("subscribers map is nil")
		}
	})
}

func TestAddSubscriber(t *testing.T) {
	t.Run("adds subscriber to map", func(t *testing.T) {
		server := FeedboardWebsocketServerBuilder.New(make(chan *zkillboard.Killmail), make(chan zkillboard.KillmailCollection))
		sub := FeedboardSubscriberBuilder.NewSubscriber(func() {})

		server.addSubscriber(&sub)

		if _, ok := server.subscribers[&sub]; !ok {
			t.Error("subscriber not added to map")
		}
	})
}

func TestRemoveSubscriber(t *testing.T) {
	t.Run("removes subscriber from map", func(t *testing.T) {
		server := FeedboardWebsocketServerBuilder.New(make(chan *zkillboard.Killmail), make(chan zkillboard.KillmailCollection))
		sub := FeedboardSubscriberBuilder.NewSubscriber(func() {})
		server.addSubscriber(&sub)

		server.removeSubscriber(&sub)

		if _, ok := server.subscribers[&sub]; ok {
			t.Error("subscriber still in map")
		}
	})
}

func TestSendTemplate(t *testing.T) {
	t.Run("sends to all subscribers", func(t *testing.T) {
		server := FeedboardWebsocketServerBuilder.New(make(chan *zkillboard.Killmail), make(chan zkillboard.KillmailCollection))
		sub := FeedboardSubscriberBuilder.NewSubscriber(func() {})
		server.addSubscriber(&sub)

		server.sendTemplate([]byte("test message"))

		select {
		case msg := <-sub.MessagesChannel:
			if string(msg) != "test message" {
				t.Errorf("got %q, want %q", msg, "test message")
			}
		default:
			t.Error("no message received")
		}
	})

	t.Run("does not call close slow when channel has space", func(t *testing.T) {
		closeCalled := false
		closeFunc := func() { closeCalled = true }
		server := FeedboardWebsocketServerBuilder.New(make(chan *zkillboard.Killmail), make(chan zkillboard.KillmailCollection))

		sub := FeedboardSubscriberBuilder.NewSubscriber(closeFunc)
		server.addSubscriber(&sub)

		server.sendTemplate([]byte("test"))

		if closeCalled {
			t.Error("CloseSlowFunc should not be called when channel has space")
		}
	})

	t.Run("calls close slow on full channel", func(t *testing.T) {
		server := FeedboardWebsocketServerBuilder.New(make(chan *zkillboard.Killmail), make(chan zkillboard.KillmailCollection))
		closeCalledChan := make(chan bool)
		closeFunc := func() { closeCalledChan <- true }
		subChan := make(chan []byte)

		sub := feedboardSubscriber{
			MessagesChannel: subChan,
			CloseSlowFunc:   closeFunc,
		}
		server.addSubscriber(&sub)
		server.sendTemplate([]byte("test"))
		select {
		case <-closeCalledChan:
		case <-time.After(10 * time.Millisecond):
			t.Fatal("timeout waiting for channel")
		}
	})
}
