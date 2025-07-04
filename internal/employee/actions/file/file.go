package file

import (
	"fmt"
	"os"
)

const (
	tmpFile = "andai-%d-*.md"
)

func BuildPromptTextTmpFile(content string) (string, error) {
	tempFile, err := os.CreateTemp("", fmt.Sprintf(tmpFile, 1))
	if err != nil {
		return "", err
	}
	//log.Printf("Created temporary file: %q", tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}
	err = tempFile.Close()

	return tempFile.Name(), err
}

func GetContents(filename string) (string, error) {
	file, err := os.ReadFile(filename) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to read file %q err: %v", filename, err)
	}
	return string(file), nil
}
