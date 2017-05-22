package config

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)

type RingpopSuite struct {
	*require.Assertions
	suite.Suite
}

func TestRingpopSuite(t *testing.T) {
	suite.Run(t, new(RingpopSuite))
}

func (s *RingpopSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *RingpopSuite) TestHostsMode() {
	var cfg Ringpop
	fmt.Println(getHostsConfig())
	err := yaml.Unmarshal([]byte(getHostsConfig()), &cfg)
	s.Nil(err)
	s.Equal("test", cfg.Name)
	s.Equal(BootstrapModeHosts, cfg.BootstrapMode)
	s.Equal([]string{"127.0.0.1:1111"}, cfg.BootstrapHosts)
	s.Equal(time.Second*30, cfg.MaxJoinDuration)
	cfg.validate()
	s.Nil(err)
	f, err := cfg.NewFactory()
	s.Nil(err)
	s.NotNil(f)
}

func (s *RingpopSuite) TestFileMode() {
	var cfg Ringpop
	err := yaml.Unmarshal([]byte(getJsonConfig()), &cfg)
	s.Nil(err)
	s.Equal("test", cfg.Name)
	s.Equal(BootstrapModeFile, cfg.BootstrapMode)
	s.Equal("/tmp/file.json", cfg.BootstrapFile)
	s.Equal(time.Second*30, cfg.MaxJoinDuration)
	err = cfg.validate()
	s.Nil(err)
	f, err := cfg.NewFactory()
	s.Nil(err)
	s.NotNil(f)
}

func (s *RingpopSuite) TestInvalidConfig() {
	var cfg Ringpop
	s.NotNil(cfg.validate())
	cfg.Name = "test"
	s.NotNil(cfg.validate())
	cfg.BootstrapMode = BootstrapModeNone
	s.NotNil(cfg.validate())
	_, err := parseBootstrapMode("unknown")
	s.NotNil(err)
}

func getJsonConfig() string {
	return `name: "test"
bootstrapMode: "file"
bootstrapFile: "/tmp/file.json"
maxJoinDuration: 30s`
}

func getHostsConfig() string {
	return `name: "test"
bootstrapMode: "hosts"
bootstrapHosts: ["127.0.0.1:1111"]
maxJoinDuration: 30s`
}
