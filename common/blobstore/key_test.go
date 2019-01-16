package blobstore

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type KeySuite struct {
	*require.Assertions
	suite.Suite
}

func TestKeySuite(t *testing.T) {
	suite.Run(t, new(KeySuite))
}

func (s *KeySuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *KeySuite) TestNewKey() {
	testCases := []struct{
		extension string
		pieces []string
		expectError bool
		builtKey string
	}{
		{
			extension: "ext",
			pieces: []string{},
			expectError: true,
		},
		{
			extension: "ext",
			pieces: []string{"1", "2", "3", "4", "5"},
			expectError: true,
		},
		{
			extension: "invalid_extension.",
			pieces: []string{"foo", "bar"},
			expectError: true,
		},
		{
			extension: "ext",
			pieces: []string{"invalid=piece"},
			expectError: true,
		},
		{
			extension: longString(60),
			pieces: []string{longString(60), longString(60), longString(60), longString(60)},
			expectError: true,
		},
		{
			extension: "ext",
			pieces: []string{"valid", "set", "of", "pieces"},
			expectError: false,
			builtKey: "valid_set_of_pieces.ext",
		},
	}

	for _, tc := range testCases {
		key, err := NewKey(tc.extension, tc.pieces...)
		if tc.expectError {
			s.Error(err)
			s.Nil(key)
		} else {
			s.NoError(err)
			s.NotNil(key)
			s.Equal(tc.builtKey, key.String())
		}
	}
}

func longString(length int) string {
	var result string
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}

