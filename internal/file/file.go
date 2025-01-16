package file

import (
	"os"
)

type File struct {
	Handle *os.File
}

func NewFile(path string) (*os.File, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return handle, nil
}

func CreateOutputFile(path string) (*os.File, error) {
	outfile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return outfile, nil
}

func CloseHandle(file *os.File) error {
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
