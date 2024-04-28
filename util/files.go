package util

import (
	"fmt"
	"io"
	"os"
)

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("can't create a file %s: %v", filename, err)
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return fmt.Errorf("can't write to a file %s: %v", filename, err)
	}
	return file.Sync()
}

func RemoveFile(filename string) error {
	err := os.Remove(filename)
	return err
}
