package cio

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileIO struct{}

func NewFileIO() *FileIO {
	return &FileIO{}
}
func (f *FileIO) Write(b []byte, path, filename string) error {
	os.MkdirAll(path, 0700)
	return ioutil.WriteFile(filepath.Join(path, filename), b, 0600)
}

func (f *FileIO) Read(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
