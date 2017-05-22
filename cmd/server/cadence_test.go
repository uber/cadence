package main

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CadenceSuite struct {
	*require.Assertions
	suite.Suite
}

func TestCadenceSuite(t *testing.T) {
	suite.Run(t, new(CadenceSuite))
}

func (s *CadenceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *CadenceSuite) TestIsValidService() {
	s.True(isValidService("history"))
	s.True(isValidService("matching"))
	s.True(isValidService("frontend"))
	s.False(isValidService("cadence-history"))
	s.False(isValidService("cadence-matching"))
	s.False(isValidService("cadence-frontend"))
	s.False(isValidService("foobar"))
}

func (s *CadenceSuite) TestPath() {
	s.Equal("foo/bar", path("foo", "bar"))
}
