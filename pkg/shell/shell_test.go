package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ShellSuite struct {
	specification *ShellSpecification
	suite.Suite
}

func (suite *ShellSuite) SetupSuite() {
	suite.specification = NewShellSpecification()
}

func (suite *ShellSuite) TestNewShellSpecification() {
	assert.Equal(suite.T(), -1, suite.specification.Cols, "should be equal")
	assert.Equal(suite.T(), -1, suite.specification.Rows, "should be equal")
	assert.Equal(suite.T(), 5, suite.specification.MinDelay, "should be equal")
	assert.Equal(suite.T(), "/bin/bash", suite.specification.Command, "should be equal")
	assert.Equal(suite.T(), "/tmp", suite.specification.Cwd, "should be equal")
	assert.Equal(suite.T(), "/tmp/recording.yaml", suite.specification.OutputPath, "should be equal")
	assert.Equal(suite.T(), make([]string, 0), suite.specification.Env, "should be equal")
}

// Run the test suite
func TestShellSuite(t *testing.T) {
	suite.Run(t, new(ShellSuite))
}