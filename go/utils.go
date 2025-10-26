package main

import (
	"io"
	"os"
)

func readJsonFile(filePath string) ([]byte, error) {
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
