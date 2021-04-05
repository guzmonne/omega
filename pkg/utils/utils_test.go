package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func cleanup (path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

func TestTouch(t *testing.T) {
	testFile := "/tmp/test.txt"
	testContent := "test"

	// Make sure the test file does not exist.
	if err := cleanup(testFile); err != nil {
		t.Fatalf(err.Error())
	}
	// Remove the test file when the test finishes.
	defer cleanup(testFile)

	// should create an empty file if one does not exist.
	if err := Touch(testFile, testContent); err != nil {
		t.Fatalf("should not have thrown error:\n%s", err.Error())
	}

	// should write the provided content to the file
	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("should not have thrown error:\n%s", err.Error())
	}
	if string(content) != testContent {
		t.Fatalf("content = %s; expected %s", string(content), testContent)
	}

}