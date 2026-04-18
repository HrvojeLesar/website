package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	eveonline "github.com/HrvojeLesar/website/eve_online"
	"github.com/HrvojeLesar/website/eve_online/feedboard"
	"github.com/HrvojeLesar/website/eve_online/zkillboard"
	"github.com/go-co-op/gocron"
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

	zkillboardR2Z2 := zkillboard.Zkillboard.NewZkillboardR2Z2(zkillboard.Zkillboard.DefaultKillmailFilterFunc)
	zkillboardR2Z2.Start()

	feedsCacheChannel := make(chan *zkillboard.Killmail, 10)
	cacheUpdateChannel := make(chan zkillboard.KillmailCollection, 10)
	feedCache := zkillboard.FiftyFiftyFiftyFeeds.NewCache(feedsCacheChannel, eveonline.KILLMAILCOUNT, cacheUpdateChannel)
	feedCache.StartListening()

	websocketListenerChannel := make(chan *zkillboard.Killmail, 10)
	killmailPropagator := zkillboard.KillmailPropagator.New(zkillboardR2Z2.KillMailsChan)
	killmailPropagator.AddListenerChannel(websocketListenerChannel)
	killmailPropagator.AddListenerChannel(feedsCacheChannel)
	killmailPropagator.Start()

	websocketServer := feedboard.FeedboardWebsocketServerBuilder.New(websocketListenerChannel, cacheUpdateChannel)
	websocketServer.StartKillmailListener()

	serveHandler := NewServeHandler(&feedCache)

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(24).Hours().Do(func() {
		slog.Info("Fetching killmails")
		feedCache.FetchKillmails()
	})

	scheduler.Every(5).Minutes().Do(serveHandler.PeriodicDocRerender)

	http.HandleFunc("/", serveHandler.Handle)
	http.HandleFunc("/feedboard-subscribe", websocketServer.SubscribeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("Listening on http://localhost%s\n", port())
	scheduler.StartAsync()
	err := http.ListenAndServe(port(), nil)
	scheduler.Stop()
	if err != nil {
		log.Panic(err)
	}
}
