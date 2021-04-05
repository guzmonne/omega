package utils

import (
	"math/rand"
	"os"
	"time"

	"github.com/oklog/ulid"
)

func Touch(filePath string, content string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// If the file does not exist we create it
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		// Write the provided content
		if _, err := file.Write([]byte(content)); err != nil {
			return err
		}
		defer file.Close()
	} else {
		// We touch the file by updating its access and modification times
		currentTime := time.Now().Local()
		if err := os.Chtimes(filePath, currentTime, currentTime); err != nil {
			return err
		}
	}

	return nil
}

func ULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}