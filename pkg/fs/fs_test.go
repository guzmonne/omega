package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"gux.codes/omega/pkg/utils"
)

type FSSuite struct {
	suite.Suite
	fs 		FS
	root	string
	file	string
}

func (suite *FSSuite) SetupSuite() {
	suite.root = "/tmp/test"
	suite.file = suite.root + "/example.txt"

	suite.cleanup()
	if err := os.MkdirAll(suite.root + "/node_modules", 0755); err != nil {
		suite.Failf("SetupSuite: %s", err.Error())
	}
	if err := utils.Touch(suite.file, ""); err != nil {
		suite.cleanup()
		suite.Failf("SetupSuite: %s", err.Error())
	}
	fs, err := NewFS(suite.root)
	if err != nil {
		suite.cleanup()
		suite.Failf("SetupSuite: %s", err.Error())
	}
	suite.fs = fs
}

func (suite *FSSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *FSSuite) cleanup() {
	if err := os.RemoveAll("/tmp/test"); err != nil {
		suite.Failf("cleanup: %s", err.Error())
	}
}

func (suite *FSSuite) TestFS() {
	suite.Run("should avoid node_modules folder", func() {
		suite.Equal([]string{suite.file}, utils.MapStringTimeKeys(suite.fs.Paths))
	})

	suite.Run("should pick up modifications on Refresh", func() {
		utils.Touch(suite.file, "")
		suite.fs.Refresh()
		suite.Equal([]string{suite.file}, utils.MapStringTimeKeys(suite.fs.Paths))
		suite.Equal(true, suite.fs.Modified)
	})

	suite.Run("should return false if everything stays equal", func() {
		suite.fs.Refresh()
		suite.Equal([]string{suite.file}, utils.MapStringTimeKeys(suite.fs.Paths))
		suite.Equal(false, suite.fs.Modified)
	})

	suite.Run("should detect new files", func() {
		utils.Touch(suite.file + "new", "")
		suite.fs.Refresh()
		suite.Equal([]string{suite.file, suite.file + "new"}, utils.MapStringTimeKeys(suite.fs.Paths))
		suite.Equal(true, suite.fs.Modified)
	})
}

// Run the test suite
func TestFSSuite(t *testing.T) {
	suite.Run(t, new(FSSuite))
}