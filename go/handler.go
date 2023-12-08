package main

import (
	"html/template"
	"net/http"
)

type ServeHandler struct {
	Esi *Esi
}

func newServeHandler(esi *Esi) *ServeHandler {
	return &ServeHandler{
		Esi: esi,
	}
}

func (s ServeHandler) handle(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/_index.html"))
	sectionWrapper := newSections(s.Esi.Killmails)
	err := tmpl.Execute(w, sectionWrapper)
	if err != nil {
		panic(err)
	}
}
