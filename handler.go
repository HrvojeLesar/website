package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/HrvojeLesar/website/eve_online/zkillboard"
	"github.com/HrvojeLesar/website/templates"
)

type ServeHandler struct {
	mainTemplate *template.Template
	feedCache    zkillboard.FiftyFiftyFiftyFeedsCache

	executedTemplate bytes.Buffer
}

func NewServeHandler(feedCache zkillboard.FiftyFiftyFiftyFeedsCache) *ServeHandler {
	sh := ServeHandler{
		feedCache: feedCache,
	}
	sh.makeTemplate()
	return &sh
}

func (sh *ServeHandler) makeTemplate() {
	sh.mainTemplate = template.Must(template.ParseFS(templates.HTMLTemplates, "_index.html", "feedboard.html", "feedboard_item.html", "feedboard_isk_total.html"))
	err := sh.executeTemplate(&sh.executedTemplate)
	if err != nil {
		log.Println(err)
	}
}

func (sh *ServeHandler) executeTemplate(w io.Writer) error {
	sectionWrapper := newSections(sh.feedCache.Killmails())
	err := sh.mainTemplate.Execute(w, sectionWrapper)
	if err != nil {
		return err
	}
	return nil
}

func (sh *ServeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if sh.feedCache.IsDirty() {
		sh.executedTemplate.Reset()
		sh.makeTemplate()
		sh.feedCache.SetNotDirty()
	}

	_, err := w.Write(sh.executedTemplate.Bytes())
	if err != nil {
		panic(err)
	}
}

func (sh *ServeHandler) PeriodicDocRerender() {
	log.Println("Periodic Rerender")
	sh.executedTemplate.Reset()
	err := sh.executeTemplate(&sh.executedTemplate)
	if err != nil {
		log.Println(err)
	}
}
