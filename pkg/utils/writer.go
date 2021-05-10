package utils

import (
	"io/ioutil"
	"os"
)

// Writer is an interface to simplify the mocking of ioutil functions
type Writer interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// FileWriter is an implementation of the Writer interface
type FileWriter struct {}

// WriteFile exposes the ioutil.WriteFile function
func (FileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

// FWriter is an exported global var of the utils package.
var FWriter FileWriter

// init configures the value of the FWriter global value
func init() {
	FWriter = FileWriter{}
}