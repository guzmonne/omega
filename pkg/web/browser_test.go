package web

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BrowserSuite struct {
	suite.Suite
}

func (suite *BrowserSuite) SetupSuite() {
	// TODO
}

func (suite *BrowserSuite) Test() {
	// TODO
}

// Run the test suite
func TestBrowserSuite(t *testing.T) {
	suite.Run(t, new(BrowserSuite))
}