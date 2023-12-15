package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
)

type ServeHandler struct {
	mainTemplate *template.Template
	Esi          *Esi

	executedTemplate bytes.Buffer
	mutex            sync.Mutex
}

func NewServeHandler(esi *Esi) *ServeHandler {
	sh := ServeHandler{
		Esi: esi,
	}
	sh.makeTemplate()
	sh.listenForKillmailUpdates()
	return &sh
}

func (sh *ServeHandler) listenForKillmailUpdates() {
	go func() {
		for {
			killmails := <-sh.Esi.TemplateCacheChan
			sh.mutex.Lock()
			defer sh.mutex.Unlock()

			sh.executedTemplate.Reset()
			err := sh.executeTemplate(&sh.executedTemplate, killmails)
			if err != nil {
				log.Println(err)
				continue
			}
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

func (sh *ServeHandler) executeTemplate(w io.Writer, killmails []FeedboardKillmail) error {
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
	err := sh.executeTemplate(&sh.executedTemplate, nil)
	if err != nil {
		log.Println(err)
	}
}
