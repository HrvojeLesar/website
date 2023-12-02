package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

type Section struct {
	Title       string `json:"Title"`
	Subsections []Subsection
}

type Subsection struct {
	Title     SubsectionTitle  `json:"title"`
	ShortText *string          `json:"short_text"`
	MainText  *string          `json:"main_text"`
	NoteText  *string          `json:"note_text"`
	Image     *string          `json:"image"`
	Links     []SubsectionLink `json:"links"`
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
	Title string  `json:"title"`
	Text  string  `json:"text"`
	Url   *string `json:"url"`
	Icon  string  `json:"icon"`
}

func readSections() []Section {
	sections := []Section{}
	fileBytes, err := readWebsiteDocsFile("website_docs/sections.json")
	if err != nil {
		log.Println(err)
		return sections
	}

	err = json.Unmarshal(fileBytes, &sections)
	if err != nil {
		log.Println(err)
		return sections
	}

	return sections
}

func readWebsiteDocsFile(filePath string) ([]byte, error) {
	jsonFile, err := os.Open(filePath)
	defer jsonFile.Close()
	if err != nil {
		return nil, err
	}

	fileBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func readSubsections(sections []Section) []Section {
	for idx := range sections {
		section := &sections[idx]
		section.Subsections = make([]Subsection, 0)
		sectionDirPath := fmt.Sprintf("website_docs/%s", section.Title)
		sectionDir, err := os.Open(sectionDirPath)
		defer sectionDir.Close()
		if err != nil {
			log.Println(err)
			continue
		}
		files, err := sectionDir.ReadDir(0)
		if err != nil {
			log.Println(err)
			continue
		}
		sort.SliceStable(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})
		for _, file := range files {
			filePath := fmt.Sprintf("%s/%s", sectionDirPath, file.Name())
			fileBytes, err := readWebsiteDocsFile(filePath)
			if err != nil {
				log.Println(err)
				continue
			}
			subsection := Subsection{}
			err = json.Unmarshal(fileBytes, &subsection)
			if err != nil {
				log.Println(err)
				continue
			}
			section.Subsections = append(section.Subsections, subsection)
		}
	}

	return sections
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
