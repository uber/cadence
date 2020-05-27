package shard

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ScannerSuite struct {
	*require.Assertions
	suite.Suite
}

func TestScannerSuite(t *testing.T) {
	suite.Run(t, new(ScannerSuite))
}

func (s *ScannerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *ScannerSuite) TestScan_Failure_FirstIteratorError() {
	// TODO: setup the mocks and write tests...
}

