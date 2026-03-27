package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
)

type Icon struct {
	Name         string `json:"name"`
	FileName     string `json:"file"`
	FileContents template.HTML
}

func loadIcons() map[string]Icon {
	icons := make([]Icon, 0)
	iconsMap := make(map[string]Icon)
	fileBytes, err := readJsonFile("link_icons/icons.json")
	if err != nil {
		log.Println(err)
		return iconsMap
	}

	err = json.Unmarshal(fileBytes, &icons)
	if err != nil {
		log.Println(err)
		return iconsMap
	}

	for idx := range icons {
		icon := &icons[idx]

		bytes, err := os.ReadFile(fmt.Sprintf("link_icons/%s", icon.FileName))
		if err != nil {
			log.Println(err)
			continue
		}

		iconFileContents := string(bytes)
		icon.FileContents = template.HTML(iconFileContents)
		iconsMap[icon.Name] = *icon
	}

	return iconsMap
}
