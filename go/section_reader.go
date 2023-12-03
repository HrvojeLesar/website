package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	Image     *Image           `json:"image"`
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
	Title    string  `json:"title"`
	Text     string  `json:"text"`
	Url      *string `json:"url"`
	Icon     string  `json:"icon"`
	IconFile *Icon
}

type SectionsWrapper struct {
	Sections []Section
}

type Image struct {
	Src *string `json:"src"`
	Alt *string `json:"alt"`
	Url *string `json:"url"`
}

func newSenctions() SectionsWrapper {
	sections := SectionsWrapper{}
	sections.readSections()
	sections.applyIcons()
	return sections
}

func (s *SectionsWrapper) readSections() {
	fileBytes, err := readJsonFile("website_docs/sections.json")
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(fileBytes, &s.Sections)
	if err != nil {
		log.Println(err)
		return
	}

	s.readSubsections()
}

func (s *SectionsWrapper) readSubsections() {
	for idx := range s.Sections {
		section := &s.Sections[idx]
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
			fileBytes, err := readJsonFile(filePath)
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
}

func (s *SectionsWrapper) applyIcons() {
	icons := loadIcons()
	for sectionIdx := range s.Sections {
		section := &s.Sections[sectionIdx]
		for subsectionIdx := range section.Subsections {
			subsection := &section.Subsections[subsectionIdx]
			for linksIdx := range subsection.Links {
				link := &subsection.Links[linksIdx]
				icon := icons[link.Icon]
				link.IconFile = &icon
			}
		}
	}
}
