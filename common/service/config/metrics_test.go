package config

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	"testing"
)

type MetricsSuite struct {
	*require.Assertions
	suite.Suite
}

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func (s *MetricsSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *MetricsSuite) TestStatsd() {
	statsd := &Statsd{
		HostPort: "127.0.0.1:8125",
		Prefix:   "testStatsd",
	}

	config := new(Metrics)
	config.Statsd = statsd
	scope := config.NewScope()
	s.NotNil(scope)
}

func (s *MetricsSuite) TestM3() {
	m3 := &m3.Configuration{
		HostPort: "127.0.0.1:8125",
		Service:  "testM3",
		Env:      "devel",
	}
	config := new(Metrics)
	config.M3 = m3
	scope := config.NewScope()
	s.NotNil(scope)
}

func (s *MetricsSuite) TestNoop() {
	config := &Metrics{}
	scope := config.NewScope()
	s.Equal(tally.NoopScope, scope)
}
