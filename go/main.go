package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	KILLMAILCOUNT int = 10
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	feedboardWebsockerServer := newFeedboardWebsocketServer()
	feedboardWebsockerServer.KillmailListener()

	esi := newEsi(KILLMAILCOUNT, feedboardWebsockerServer.KillmailChan)

	zkm := NewZkillWebsocketManager(esi.handleWebsocketKillmail, ZkillWebsocketFilter{Action: "sub", Channel: "corporation:98684728"})
	zkm.Run()
	serveHandler := NewServeHandler(esi)

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(24).Hours().Do(func() {
		log.Println("Fetching killmails")
		err := esi.fetchFiftyFiftyFiftyFeeds()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Fetching successful")
		}
	})

	scheduler.Every(5).Minutes().Do(serveHandler.PeriodicDocRerender)

	http.HandleFunc("/", serveHandler.Handle)
	http.HandleFunc("/feedboard-subscribe", feedboardWebsockerServer.SubscribeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Listening on http://localhost:3000")
	scheduler.StartAsync()
	err := http.ListenAndServe(":3000", nil)
	scheduler.Stop()
	if err != nil {
		log.Panic(err)
	}
}
