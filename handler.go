package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"

	eveonline "github.com/HrvojeLesar/website/eve_online"
)

type ServeHandler struct {
	mainTemplate *template.Template

	executedTemplate bytes.Buffer
	mutex            sync.Mutex
}

func NewServeHandler(esi *eveonline.Esi) *ServeHandler {
	sh := ServeHandler{}
	sh.makeTemplate()
	sh.listenForKillmailUpdates()
	return &sh
}

// Mark template as dirty and rerender it when actually serving
// Be lazy and only do work when requested
func (sh *ServeHandler) listenForKillmailUpdates() {
	go func() {
		for {
			killmails := <-sh.Esi.TemplateCacheChan
			sh.mutex.Lock()

			sh.executedTemplate.Reset()
			err := sh.executeTemplate(&sh.executedTemplate, killmails)
			if err != nil {
				log.Println(err)
				sh.mutex.Unlock()
				continue
			}
			sh.mutex.Unlock()
		}
	}()
}

func (sh *ServeHandler) makeTemplate() {
	sh.mainTemplate = template.Must(template.ParseFiles("templates/_index.html", "templates/feedboard.html", "templates/feedboard_item.html"))
	err := sh.executeTemplate(&sh.executedTemplate, nil)
	if err != nil {
		log.Println(err)
	}
}

func (sh *ServeHandler) executeTemplate(w io.Writer, killmails []eveonline.FeedboardKillmail) error {
	sectionWrapper := newSections(killmails)
	err := sh.mainTemplate.Execute(w, sectionWrapper)
	if err != nil {
		return err
	}
	return nil
}

func (sh *ServeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(sh.executedTemplate.Bytes())
	if err != nil {
		panic(err)
	}
}

func (sh *ServeHandler) PeriodicDocRerender() {
	log.Println("Periodic Rerender")
	sh.executedTemplate.Reset()
	err := sh.executeTemplate(&sh.executedTemplate, sh.Esi.Killmails)
	if err != nil {
		log.Println(err)
	}
}
