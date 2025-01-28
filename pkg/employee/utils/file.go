package utils

import (
	"fmt"
	"log"
	"os"
)

func BuildPromptTextTmpFile(content string) (string, error) {
	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, 1))
	if err != nil {
		return "", err
	}
	log.Printf("Created temporary file: %q", tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), nil
}
