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

package store

import (
	"context"
	"testing"

	"github.com/uber/cadence/common/blobstore"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/blobstore/filestore"
	"github.com/uber/cadence/common/config"
)

type BlobstoreWriterSuite struct {
	suite.Suite
	*require.Assertions
	blobstoreWriter *blobstoreWriter
	uuid            string
	extension       Extension
	outputDir       string
	blobstoreClient blobstore.Client
}

func TestBlobstoreWriterSuite(t *testing.T) {
	suite.Run(t, new(BlobstoreWriterSuite))
}

func (s *BlobstoreWriterSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.uuid = "test-uuid"
	s.extension = Extension("test")
	s.outputDir = s.T().TempDir()

	cfg := &config.FileBlobstore{
		OutputDirectory: s.outputDir,
	}
	var err error
	s.blobstoreClient, err = filestore.NewFilestoreClient(cfg)
	require.NoError(s.T(), err)

	s.blobstoreWriter = NewBlobstoreWriter(s.uuid, s.extension, s.blobstoreClient, 10).(*blobstoreWriter)
}

func (s *BlobstoreWriterSuite) TestBlobstoreWriter() {
	// Example test data
	testData := "test-data"

	// Add data to the writer
	err := s.blobstoreWriter.Add(testData)
	s.NoError(err)

	// Flush the writer to write data to the blobstore
	err = s.blobstoreWriter.Flush()
	s.NoError(err)

	// Retrieve the keys of flushed data
	flushedKeys := s.blobstoreWriter.FlushedKeys()
	s.NotNil(flushedKeys)

	// Read back the data from the blobstore
	key := pageNumberToKey(s.uuid, s.extension, flushedKeys.MinPage)
	req := &blobstore.GetRequest{Key: key}
	ctx := context.Background()
	resp, err := s.blobstoreClient.Get(ctx, req)
	s.NoError(err)

	// Verify the contents
	s.Equal(testData, string(resp.Blob.Body))
}

// Additional test cases can be added here
