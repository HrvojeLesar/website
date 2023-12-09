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

	esi := newEsi(KILLMAILCOUNT)

	zkm := NewZkillWebsocketManager(esi.handleWebsocketKillmail, ZkillWebsocketFilter{Action: "sub", Channel: "corporation:98684728"})
	zkm.Run()

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

	serveHandler := newServeHandler(esi)
	http.HandleFunc("/", serveHandler.handle)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Listening on http://localhost:3000")
	scheduler.StartAsync()
	err := http.ListenAndServe(":3000", nil)
	scheduler.Stop()
	if err != nil {
		log.Panic(err)
	}
}
