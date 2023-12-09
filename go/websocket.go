package main

import (
	"context"
	"errors"
	"log"
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
