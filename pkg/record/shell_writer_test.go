package record

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ShellWriterSuite struct {
	suite.Suite
	writer *ShellWriter
	specification *ShellSpecification
}

func (suite *ShellWriterSuite) SetupSuite() {
	suite.specification = NewShellSpecification()
	suite.writer = NewShellWriter(suite.specification)
}

func (suite *ShellWriterSuite) TestNewShellWriter() {
	assert.Equal(suite.T(), suite.specification, suite.writer.specification, "should be equal")
	assert.Equal(suite.T(), make([]Record, 0), suite.writer.records, "should be equal")
	assert.WithinDuration(suite.T(), time.Now(), suite.writer.timestamp, time.Millisecond * 1.0, "should be equal")
}

func (suite *ShellWriterSuite) TestShellWriter() {
	suite.Run(".Write()", func() {
		_, err := suite.writer.Write([]byte("something"))
		// should throw no errors
		assert.NoError(suite.T(), err)
		// first record should have a delay of 0
		assert.Equal(suite.T(), 0, suite.writer.records[0].Delay, "should be equal")
		// if the time between two Write invocations is less than MIN_DELAY it should
		// append the content to the previous record instead of creating a new one.
		suite.writer.Write([]byte(" "))
		suite.writer.Write([]byte("else"))
		assert.Equal(suite.T(), "something else", suite.writer.records[0].Content, "should overwrite previous record content")
		// should allow the creation of more records
		time.Sleep(time.Millisecond * time.Duration(suite.specification.MinDelay + 1))
		suite.writer.Write([]byte("1"))
		time.Sleep(time.Millisecond * time.Duration(suite.specification.MinDelay + 1))
		suite.writer.Write([]byte("2"))
		time.Sleep(time.Millisecond * time.Duration(suite.specification.MinDelay + 1))
		suite.writer.Write([]byte("3"))
		assert.Equal(suite.T(), 4, len(suite.writer.records), "should have 4 records")
	})
}

// Run the test suite
func TestShellWriterSuite(t *testing.T) {
	suite.Run(t, new(ShellWriterSuite))
}