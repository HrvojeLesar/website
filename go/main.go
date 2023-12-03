package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	tmpl := template.Must(template.ParseFiles("templates/_index.html"))
	sectionWrapper := newSenctions()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, sectionWrapper.Sections)
		if err != nil {
			panic(err)
		}
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Listening on http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Panic(err)
	}
}
