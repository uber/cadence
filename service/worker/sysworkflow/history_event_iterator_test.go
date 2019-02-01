package sysworkflow

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type HistoryEventIteratorSuite struct {
	*require.Assertions
	suite.Suite
}

func TestHistoryEventIteratorSuite(t *testing.T) {
	suite.Run(t, new(HistoryEventIteratorSuite))
}

func (s *HistoryEventIteratorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *HistoryEventIteratorSuite) TestTest() {
	s.True(true)
}
