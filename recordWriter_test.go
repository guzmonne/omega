package main

import (
	"testing"
	"time"
)

func TestRecordWriterWrite(t *testing.T) {
	writer := NewRecordWriter()

	// Should not fail when trying to write something
	_, err := writer.Write([]byte("something"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// If it's the first record it should be added with a delay equal to zero.
	if len(writer.records) == 1 {
		records := writer.records
		record := records[0]
		if record.Delay != 0 {
			t.Fatalf("record.Delay = %d; expected 0", record.Delay)
		}
	} else {
		t.Fatalf("len(writer.records) = %d; expected 1", len(writer.records))
	}

	// If the time between each invocation is less than MIN_DELAY it should append
	// the content to the previous record instead of creating a new record.
	writer.Write([]byte(" "))
	writer.Write([]byte("else"))
	record := writer.records[len(writer.records) - 1]
	if record.Content != "something else" {
		t.Fatalf("record.Content = %s; expected 12", record.Content)
	}

	// should continue adding more records as the `Write` function is called.
	time.Sleep(MIN_DELAY + 1000)
	writer.Write([]byte("something"))
	time.Sleep(MIN_DELAY + 1000)
	writer.Write([]byte("something"))
	time.Sleep(MIN_DELAY + 1000)
	writer.Write([]byte("something"))
	if len(writer.records) != 4 {
		t.Fatalf("len(writer.records) = %d; expected 4", len(writer.records))
	}
}