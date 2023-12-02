package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Section struct {
	Title string `json:"Title"`
}

type Subsection struct {
	ShortText *string          `json:"short_text"`
	MainText  *string          `json:"main_text"`
	NoteText  *string          `json:"note_text"`
	Links     []SubsectionLink `json:"links"`
	Image     *string          `json:"image"`
}

type SubsectionTitle struct {
	Title string  `json:"title"`
	Url   *string `json:"url"`
}

type TemplateData struct {
	Sections    []Section
	Subsections []Subsection
}

type SubsectionLink struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Icon  string `json:"icon"`
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/_index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		section := Section{Title: "Projects"}
		subsection := Subsection{}
		tmplData := TemplateData{
			Sections:    []Section{section},
			Subsections: []Subsection{subsection},
		}
		err := tmpl.Execute(w, tmplData)
		if err != nil {
			panic(err)
		}
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":3000", nil)
	fmt.Println("Listening on http://localhost:3000")
}
