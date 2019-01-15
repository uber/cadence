// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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

func (s *BlobSuite) TestCompress_Fail_AlreadyCompressed() {
	b := &blob{
		tags: map[string]string{compressionTag: "exists"},
	}
	cBlob, err := b.Compress()
	s.Error(err)
	s.Equal(errBlobAlreadyCompressed, err)
	s.Nil(cBlob)
}

func (s *BlobSuite) TestDecompress_Fail_AlreadyDecompressed() {
	b := &blob{}
	dBlob, err := b.Decompress()
	s.Error(err)
	s.Equal(errBlobAlreadyDecompressed, err)
	s.Nil(dBlob)
}

func (s *BlobSuite) TestDecompress_Fail_GzipFormatInvalid() {
	b := &blob{
		tags: map[string]string{compressionTag: gzipCompression},
		body: []byte("invalid gzip format"),
	}
	dBlob, err := b.Decompress()
	s.Error(err)
	s.Nil(dBlob)
}

func (s *BlobSuite) TestDecompress_Fail_InvalidCompressionType() {
	b := &blob{
		tags: map[string]string{compressionTag: "invalid"},
	}
	dBlob, err := b.Decompress()
	s.Error(err)
	s.Nil(dBlob)
}

func (s *BlobSuite) TestDecompress_Success_NoCompression() {
	b := &blob{
		tags: map[string]string{compressionTag: noCompression},
		body: []byte("non-compressed blob body"),
	}
	dBlob, err := b.Decompress()
	s.NoError(err)
	s.NotNil(dBlob)
	s.Empty(dBlob.Tags())
	s.Equal([]byte("non-compressed blob body"), dBlob.Body())
}

func (s *BlobSuite) TestCompressDecompress_Success_GzipFormat() {
	body := []byte("some random blob body")
	b := &blob{
		tags: map[string]string{"user_tag": "user_value"},
		body: body,
	}
	cBlob, err := b.Compress()
	s.NoError(err)
	s.NotNil(cBlob)
	expectedTags := map[string]string{
		"user_tag": "user_value",
		compressionTag: gzipCompression,
	}
	s.Equal(expectedTags, cBlob.Tags())
	s.NotEqual(body, cBlob.Body())

	dBlob, err := cBlob.Decompress()
	s.NoError(err)
	s.NotNil(dBlob)
	expectedTags =  map[string]string{"user_tag": "user_value"}
	s.Equal(expectedTags, dBlob.Tags())
	s.Equal(body, dBlob.Body())
}