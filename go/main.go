package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	KILLMAILCOUNT int = 10
)

func port() string {
	port, isSet := os.LookupEnv("GO_PORT")
	if isSet {
		return fmt.Sprintf(":%s", port)
	} else {
		return ":3000"
	}

}

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

	fmt.Printf("Listening on http://localhost%s\n", port())
	scheduler.StartAsync()
	err := http.ListenAndServe(port(), nil)
	scheduler.Stop()
	if err != nil {
		log.Panic(err)
	}
}
