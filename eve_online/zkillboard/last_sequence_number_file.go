package zkillboard

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type sequenceFile struct{}

var SequenceFile sequenceFile

const filePath = "last-sequence.txt"
const filePermissioins = 0644

func (sequenceFile *sequenceFile) LastSaved() (SequenceNumber, error) {
	var sequenceNumber SequenceNumber

	if !sequenceFile.fileExists() {
		return sequenceNumber, fmt.Errorf("Cannot read sequence number, file %s does not exists", filePath)
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return sequenceNumber, err
	}

	number, err := strconv.ParseUint(strings.TrimSpace(string(fileData)), 10, 64)
	if err != nil {
		return sequenceNumber, err
	}

	sequenceNumber.Sequence = number
	return sequenceNumber, nil
}

func (sequenceFile *sequenceFile) Save(sequenceNumber SequenceNumber) error {
	err := os.WriteFile(filePath, []byte(strconv.FormatUint(sequenceNumber.Sequence, 10)), filePermissioins)
	if err != nil {
		return err
	}

	return nil
}

func (sequenceFile *sequenceFile) fileExists() bool {
	_, err := os.Stat(filePath)

	return !errors.Is(err, os.ErrNotExist)
}
