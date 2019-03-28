package blob

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

func (s *BlobSuite) TestEquals() {
	testCases := []struct {
		blobA *Blob
		blobB *Blob
		equal bool
	}{
		{
			blobA: nil,
			blobB: nil,
			equal: true,
		},
		{
			blobA: NewBlob(nil, nil),
			blobB: nil,
			equal: false,
		},
		{
			blobA: NewBlob([]byte{1, 2, 3}, map[string]string{}),
			blobB: NewBlob([]byte{1, 2, 3}, map[string]string{"k1": "v1"}),
			equal: false,
		},
		{
			blobA: NewBlob(nil, map[string]string{"k1": "v1"}),
			blobB: NewBlob([]byte{1}, map[string]string{"k1": "v1"}),
			equal: false,
		},
		{
			blobA: NewBlob([]byte{1, 2, 3}, map[string]string{"k1": "v1"}),
			blobB: NewBlob([]byte{1, 2, 3, 4}, map[string]string{"k1": "v1"}),
			equal: false,
		},
		{
			blobA: NewBlob([]byte{1, 2, 3}, map[string]string{"k1": "v1"}),
			blobB: NewBlob([]byte{1, 2, 3}, map[string]string{"k1_x": "v1_x"}),
			equal: false,
		},
		{
			blobA: NewBlob([]byte{1, 2}, map[string]string{"k1": "v1", "k2": "v2"}),
			blobB: NewBlob([]byte{1, 2}, map[string]string{"k1": "v1"}),
			equal: false,
		},
		{
			blobA: NewBlob([]byte{1, 2}, map[string]string{"k1": "v1", "k2": "v2"}),
			blobB: NewBlob([]byte{1, 2}, map[string]string{"k1": "v1", "k2": "v2"}),
			equal: true,
		},
	}
	for _, tc := range testCases {
		s.Equal(tc.equal, tc.blobA.Equal(tc.blobB))
		s.Equal(tc.equal, tc.blobA.Equal(tc.blobB))
		aDeepCopy := tc.blobA.DeepCopy()
		bDeepCopy := tc.blobB.DeepCopy()
		s.Equal(tc.equal, aDeepCopy.Equal(bDeepCopy))
	}
}
