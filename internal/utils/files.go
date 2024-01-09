package utils

import (
	"io"
	"os"
	"path/filepath"
)

func CopyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func SafelyOverwriteFile(filePath string, content string) error {
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "swervo_temp")
	if err != nil {
		return err
	}

	_, err = tempFile.WriteString(content)
	if err != nil {
		return err
	}

	err = tempFile.Sync()
	if err != nil {
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), filePath)
	if err != nil {
		return err
	}

	return nil
}
