package config

import (
	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type LogSuite struct {
	*require.Assertions
	suite.Suite
}

func TestLogSuite(t *testing.T) {
	suite.Run(t, new(LoaderSuite))
}

func (s *LogSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *LogSuite) TestParseLogLevel() {
	s.Equal(logrus.DebugLevel, parseLogrusLevel("debug"))
	s.Equal(logrus.InfoLevel, parseLogrusLevel("info"))
	s.Equal(logrus.WarnLevel, parseLogrusLevel("warn"))
	s.Equal(logrus.ErrorLevel, parseLogrusLevel("error"))
	s.Equal(logrus.FatalLevel, parseLogrusLevel("fatal"))
	s.Equal(logrus.InfoLevel, parseLogrusLevel("unknown"))
}

func (s *LogSuite) TestNewLogger() {

	dir, err := ioutil.TempDir("", "config.testNewLogger")
	s.Nil(err)
	defer os.RemoveAll(dir)

	config := &Logger{
		Stdout:     true,
		Level:      "info",
		OutputFile: dir + "/test.log",
	}

	log := config.NewBarkLogger()
	s.NotNil(log)
	_, err = os.Stat(dir + "/test.log")
	s.Nil(err)
}
