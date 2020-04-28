// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package util

import (
	"github.com/uber/cadence/common/blobstore"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var blobMap = map[string]blobstore.Blob{
	"empty": {
		Body: nil,
		Tags: nil,
	},
	"empty_body": {
		Body: nil,
		Tags: map[string]string{"key_1": "value_1"},
	},
	"empty_tags": {
		Body: []byte("\"one\"\r\n\"two\"\r\n"),
		Tags: nil,
	},
	"blob_1": {
		Body: []byte("\"three\"\r\n\"four\"\r\n\"five\"\r\n"),
		Tags: map[string]string{"key1": "value1"},
	},
	"blob_2": {
		Body: []byte("\"abc\"\r\n\"def\"\r\n\"ghi\"\r\n"),
		Tags: map[string]string{"key1": "value1", "key2": "value2"},
	},
	"blob_3": {
		Body: []byte("\"dog\"\r\n\"cat\"\r\n\"fish\"\r\n"),
		Tags: map[string]string{"key1": "value1", "key2": "value2"},
	},
}

type IteratorSuite struct {
	*require.Assertions
	suite.Suite
}

func TestIteratorSuite(t *testing.T) {
	suite.Run(t, new(IteratorSuite))
}

func (s *IteratorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

//func (s *IteratorSuite) TestInitializedToEmpty() {
//	getFn := func(key string) (blobstore.Blob, error) {
//		return blobstore.Blob{}, nil
//	}
//	itr := NewIterator([]string{"key_1", "key_2", "key_3"}, getFn, []byte("\r\n"))
//	s.assertIteratorState(itr, false, nil, nil, false, true)
//}

//func (s *IteratorSuite) TestMultiBlobPagesNoErrors() {
//	getFn := func(key string) (blobstore.Blob, error) {
//		return blobMap[key], nil
//	}
//	itr := NewIterator(
//		[]string{"blob_3", "empty", "blob_2", "empty_body", "blob_1", "empty_tags"},
//		getFn,
//		[]byte("\r\n"))
//
//	s.assertIteratorState(itr, true, blobMap["blob_3"].Tags, []byte("\"dog\""), true, false)
//	s.assertIteratorState(itr, true, blobMap["blob_3"].Tags, []byte("\"cat\""), false, false)
//	s.assertIteratorState(itr, true, blobMap["blob_3"].Tags, []byte("\"fish\""), false, false)
//	s.assertIteratorState(itr, true, blobMap["blob_2"].Tags, []byte("\"abc\""), true, false)
//	s.assertIteratorState(itr, true, blobMap["blob_2"].Tags, []byte("\"def\""), false, false)
//	s.assertIteratorState(itr, true, blobMap["blob_2"].Tags, []byte("\"ghi\""), false, false)
//	s.assertIteratorState(itr, true, blobMap["blob_1"].Tags, []byte("\"three\""), true, false)
//	s.assertIteratorState(itr, true, blobMap["blob_1"].Tags, []byte("\"four\""), false, false)
//	s.assertIteratorState(itr, true, blobMap["blob_1"].Tags, []byte("\"five\""), false, false)
//	s.assertIteratorState(itr, true, nil, []byte("\"one\""), true, false)
//	s.assertIteratorState(itr, true, nil, []byte("\"two\""), false, false)
//}

//func (s *IteratorSuite) assertIteratorState(
//	itr Iterator,
//	expectedHasNext bool,
//	expectedTags map[string]string,
//	expectedValue []byte,
//	expectNewTags bool,
//	expectedError bool,
//) {
//	s.Equal(expectedHasNext, itr.HasNext())
//	s.Equal(expectedTags, itr.Tags())
//	value, newTags, err := itr.Next()
//	s.Equal(expectedValue, value)
//	s.Equal(expectNewTags, newTags)
//	if expectedError {
//		s.Error(err)
//	} else {
//		s.NoError(err)
//	}
//}
