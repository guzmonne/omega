package chrome

import (
	"fmt"
	"io/ioutil"

	"github.com/cheggaaa/pb/v3"
)

// FramesWriter writes inputs as Framess
type FramesWriter struct {
	frames [][]byte
}

// NewFramesWriter creates a new FramesWriter
func NewFramesWriter() FramesWriter {
	return FramesWriter{
		frames: make([][]byte, 0),
	}
}

// Write adds a new []byte into the `frames` slice
func (writer *FramesWriter) Write(input []byte) (int, error) {
	writer.frames = append(writer.frames, input)

	// Comply with the Writer interface
	return len(input), nil
}

// Dump writes all the frames stored into a list of files whose
// names should be specified as a fileStringTemplate.
func (writer FramesWriter) Dump(fileStringTemplate string) error {
	// Create a bar to let the user know of the dump status.
	bar := pb.StartNew(len(writer.frames))

	// Write the files to the filesystem
	for index, value := range writer.frames {
		err := ioutil.WriteFile(fmt.Sprintf(fileStringTemplate, index), value, 0644)
		if err != nil {
			return err
		}
		// Move the bar forward
		bar.Increment()
	}

	// Close the bar progress
	bar.Finish()

	return nil
}