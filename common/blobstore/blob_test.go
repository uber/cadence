package blobstore

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BlobSuite struct {
	*require.Assertions
	suite.Suite
}

func TestBlobSuite(t *testing.T) {
	suite.Run(t, new(BlobSuite))
}

func (s *BlobSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

/**
test should basically all do the following
- construct a blob
- compress the blob and make assertion about the tags
- decompress the blob assert tags match and body matches original blob


 */






func (s *BlobSuite) TestTest() {
	s.NotNil("foo")
}