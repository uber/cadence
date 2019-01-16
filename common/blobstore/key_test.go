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
	testCases := []struct {
		extension      string
		pieces         []string
		expectError    bool
		expectBuiltKey string
	}{
		{
			extension:   "ext",
			pieces:      []string{},
			expectError: true,
		},
		{
			extension:   "ext",
			pieces:      []string{"1", "2", "3", "4", "5"},
			expectError: true,
		},
		{
			extension:   "invalid_extension.",
			pieces:      []string{"foo", "bar"},
			expectError: true,
		},
		{
			extension:   "ext",
			pieces:      []string{"invalid=piece"},
			expectError: true,
		},
		{
			extension:   longString(60),
			pieces:      []string{longString(60), longString(60), longString(60), longString(60)},
			expectError: true,
		},
		{
			extension:      "ext",
			pieces:         []string{"valid", "set", "of", "pieces"},
			expectError:    false,
			expectBuiltKey: "valid_set_of_pieces.ext",
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
			s.Equal(tc.expectBuiltKey, key.String())
		}
	}
}

func (s *KeySuite) TestNewKeyFromString() {
	testCases := []struct {
		inputStr         string
		expectError      bool
		expectBuiltKey   string
		expectExtension  string
		expectNamePieces []string
	}{
		{
			inputStr:    "",
			expectError: true,
		},
		{
			inputStr:    "ext",
			expectError: true,
		},
		{
			inputStr:    "foo.bar.baz.ext",
			expectError: true,
		},
		{
			inputStr:    "1_2_3_4_5.ext",
			expectError: true,
		},
		{
			inputStr:    "1=4_5.ext",
			expectError: true,
		},
		{
			inputStr:    "foo_bar_bax.e,x,t,",
			expectError: true,
		},
		{
			inputStr:         "foo1_bar2_3baz.ext",
			expectError:      false,
			expectBuiltKey:   "foo1_bar2_3baz.ext",
			expectExtension:  "ext",
			expectNamePieces: []string{"foo1", "bar2", "3baz"},
		},
	}

	for _, tc := range testCases {
		key, err := NewKeyFromString(tc.inputStr)
		if tc.expectError {
			s.Error(err)
			s.Nil(key)
		} else {
			s.NoError(err)
			s.NotNil(key)
			s.Equal(tc.expectBuiltKey, key.String())
			s.Equal(tc.expectNamePieces, key.Pieces())
			s.Equal(tc.expectExtension, key.Extension())
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
