package main

import (
	"io"
	"os"
)

func readJsonFile(filePath string) ([]byte, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	fileBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
