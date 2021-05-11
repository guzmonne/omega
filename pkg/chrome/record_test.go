package chrome

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	mbrowser "gux.codes/omega/mocks/browser"
	mutils "gux.codes/omega/mocks/utils"
	"gux.codes/omega/pkg/browser"
)

type RecordSuite struct {
	suite.Suite
	writer mutils.Writer
}

func (suite *RecordSuite) SetupSuite() {
	suite.writer = mutils.Writer{}
	browser.Chrome = &mbrowser.BrowserHandler{}
}

func (suite *RecordSuite) TearDownSuite() {
	browser.Chrome = browser.ChromeBrowser{}
}

func (suite *RecordSuite) TestRecord() {
	// Define the parent context
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create the RecordParams object
	params := RecordParams{
		Duration: 1000,
		URL     : "https://example.com",
		Writer  : &suite.writer,
		Workers : 8,
		Width : 1920,
		Height: 1080,
	}

	// Default browser behavior
	browser.Chrome.(*mbrowser.BrowserHandler).On("NewContext", parent).Return(context.WithCancel(context.Background()))
	browser.Chrome.(*mbrowser.BrowserHandler).On("Navigate", parent, params.URL, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).Return(nil)
	browser.Chrome.(*mbrowser.BrowserHandler).On("Evaluate", parent, mock.AnythingOfType("string")).Return(nil, nil)
	browser.Chrome.(*mbrowser.BrowserHandler).On("Screenshot", parent).Return(nil, nil)
	suite.writer.On("WriteFile", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	// Run the record function
	record(parent, params)

	// Check that an appropiate number of browser contexts was created
	browser.Chrome.(*mbrowser.BrowserHandler).AssertNumberOfCalls(suite.T(), "NewContext", params.Workers)

	// Calculate the amount of frames ro record.
	framesToRecord  := math.Ceil(params.Duration * FPS / 1000)

	// Check that the screenshot function was called for each frame
	browser.Chrome.(*mbrowser.BrowserHandler).AssertNumberOfCalls(suite.T(), "Screenshot", int(framesToRecord))

	// Check that the write function was called for each frame
	suite.writer.AssertNumberOfCalls(suite.T(), "WriteFile", int(framesToRecord))

	// Check that the evaluate function was called to go to the first and last frame
	browser.Chrome.(*mbrowser.BrowserHandler).AssertCalled(suite.T(), "Evaluate", mock.Anything, `timeweb.goTo(16.667)`)
	browser.Chrome.(*mbrowser.BrowserHandler).AssertCalled(suite.T(), "Evaluate", mock.Anything, `timeweb.goTo(966.667)`)
	browser.Chrome.(*mbrowser.BrowserHandler).AssertCalled(suite.T(), "Evaluate", mock.Anything, `timeweb.goTo(983.333)`)
}

// Run the test suite
func TestRecordSuite(t *testing.T) {
	suite.Run(t, new(RecordSuite))
}