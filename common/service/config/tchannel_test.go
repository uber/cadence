package config

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/tchannel-go"
	"testing"
)

type TChannelSuite struct {
	*require.Assertions
	suite.Suite
}

func TestTChannelSuite(t *testing.T) {
	suite.Run(t, new(TChannelSuite))
}

func (s *TChannelSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *TChannelSuite) TestParseLogLevel() {
	var cfg TChannel
	cfg.LogLevel = "debug"
	s.Equal(tchannel.LogLevelDebug, cfg.getLogLevel())
	cfg.LogLevel = "info"
	s.Equal(tchannel.LogLevelInfo, cfg.getLogLevel())
	cfg.LogLevel = "warn"
	s.Equal(tchannel.LogLevelWarn, cfg.getLogLevel())
	cfg.LogLevel = "error"
	s.Equal(tchannel.LogLevelError, cfg.getLogLevel())
	cfg.LogLevel = "fatal"
	s.Equal(tchannel.LogLevelFatal, cfg.getLogLevel())
	cfg.LogLevel = ""
	s.Equal(tchannel.LogLevelWarn, cfg.getLogLevel())
}

func (s *TChannelSuite) TestFactory() {
	cfg := &TChannel{
		LogLevel:        "info",
		BindOnLocalHost: true,
	}
	f := cfg.NewFactory()
	s.NotNil(f)
	ch, _ := f.CreateChannel("test", nil)
	s.NotNil(ch)
	ch.Close()
}
