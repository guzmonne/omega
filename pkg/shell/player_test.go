package shell

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type PlaySuite struct {
	recordingPath string
	records []Record
	playOptions PlayOptions
	suite.Suite
}

func (suite *PlaySuite) cleanup() {
	// Make sure there are no existing files
	if err := os.RemoveAll(suite.recordingPath); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *PlaySuite) createRecording() {
	var file bytes.Buffer
	encoder := yaml.NewEncoder(&file)
	encoder.SetIndent(2)

	err := encoder.Encode(suite.records)
	if err != nil {
		suite.FailNow(err.Error())
	}

	err = ioutil.WriteFile(suite.recordingPath, file.Bytes(), 0644)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *PlaySuite) SetupSuite() {
	suite.recordingPath = "/tmp/recording.yml"
	suite.records = []Record{{Delay: 0, Content: "0"}, {Delay: 1, Content: "1"}}
	suite.playOptions = NewPlayOptions()
}

func (suite *PlaySuite) SetupTest() {
	suite.cleanup()
	suite.createRecording()
}

func (suite *PlaySuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *PlaySuite) TestReadRecording() {
	content, err := ReadRecording(suite.recordingPath)
	suite.NoError(err)
	suite.Equal(content, suite.records, "should be equal")
}

func (suite *PlaySuite) TestNewPlayOptions() {
	suite.Equal(1.0, suite.playOptions.SpeedFactor)
	suite.Equal(false, suite.playOptions.Silent)
	suite.Equal(-1, suite.playOptions.FrameDelay)
	suite.Equal(-1, suite.playOptions.MaxIdleTime)
}

func (suite *PlaySuite) TestAdjustFrameDelay() {
	suite.Run("should do nothing if the default options are provided", func() {
		records := AdjustFrameDelay(suite.records, suite.playOptions)
		suite.Equal(suite.records, records)
	})

	suite.Run("should apply the frameDelay option if it's not -1", func() {
		frameDelay := 100
		playOptions := NewPlayOptions()
		playOptions.FrameDelay = frameDelay
		original := []Record{{Delay: 1, Content: "1"}, {Delay: 2, Content: "2"}}
		expected := []Record{{Delay: frameDelay, Content: "1"}, {Delay: frameDelay, Content: "2"}}
		actual := AdjustFrameDelay(original, playOptions)
		suite.Equal(expected, actual)
	})

	suite.Run("should max the delay if maxIdleDelay is not -1", func() {
		maxIdleDelay := 50
		playOptions := NewPlayOptions()
		playOptions.MaxIdleTime = maxIdleDelay
		original := []Record{{Delay: 100, Content: "1"}, {Delay: 200, Content: "2"}}
		expected := []Record{{Delay: maxIdleDelay, Content: "1"}, {Delay: maxIdleDelay, Content: "2"}}
		actual := AdjustFrameDelay(original, playOptions)
		suite.Equal(expected, actual)
	})

	suite.Run("should multiply the delay by the speed factor", func() {
		speedFactor := 2.0
		playOptions := NewPlayOptions()
		playOptions.SpeedFactor = speedFactor
		original := []Record{{Delay: 100, Content: "1"}, {Delay: 200, Content: "2"}}
		expected := []Record{{Delay: int(100 * speedFactor), Content: "1"}, {Delay: int(200 * speedFactor), Content: "2"}}
		actual := AdjustFrameDelay(original, playOptions)
		suite.Equal(expected, actual)
	})
}

// Run the test suite
func TestPlaySuite(t *testing.T) {
	suite.Run(t, new(PlaySuite))
}
