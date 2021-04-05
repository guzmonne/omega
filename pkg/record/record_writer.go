package record

import (
	"time"
)

// RecordWriter writes inputs as Records
type RecordWriter struct {
	timestamp time.Time
	records []Record
}
// MIN_DELAY corresponds to the minimum time that needs to elapse to consider the
// creation of a new Record in ms.
const MIN_DELAY = 5
func NewRecordWriter() *RecordWriter {
	// Create an empty []Record
	records := make([]Record, 0)
	// Return the RecordWriter
	return &RecordWriter{time.Now(), records}
}
// Write the input as a new Record. If the time since the last Record is less
// than MIN_DELAY, then it modifies the last record appending the new bytes.
func (writer *RecordWriter) Write(input []byte) (int, error) {
	// Calculate the delay since the last record
	var delay int
	if len(writer.records) == 0 {
		record := &Record{0, string(input)}
		writer.records = append(writer.records, *record)
	} else {
		// If the delay is less than MIN_DELAY then we get the previous record
		// and update it. Else we create a new one.
		delay = int(time.Since(writer.timestamp) / 1000 / 1000)
		if delay < MIN_DELAY {
			// Get the previous Record and update its content
			previous := &writer.records[len(writer.records) - 1]
			previous.Content = previous.Content + string(input)
		} else {
			record := &Record{delay, string(input)}
			writer.records = append(writer.records, *record)
		}
	}
	// Update the writer timestamp
	writer.timestamp = time.Now()

	// Comply with the Writer interface
	return len(input), nil
}