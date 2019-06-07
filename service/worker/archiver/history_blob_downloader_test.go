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

package archiver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore/blob"
	"github.com/uber/cadence/common/mocks"
	"testing"
)

type historyBlobDownloaderSuite struct {
	suite.Suite
	blobstoreClient *mocks.BlobstoreClient
}

func TestHistoryBlobDownloderSuite(t *testing.T) {
	suite.Run(t, new(historyBlobDownloaderSuite))
}

func (h *historyBlobDownloaderSuite) SetupTest() {
	h.blobstoreClient = &mocks.BlobstoreClient{}
}

func (h *historyBlobDownloaderSuite) TearDownTest() {
	h.blobstoreClient.AssertExpectations(h.T())
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Failed_CouldNotDeserializeToken() {
	blobstoreClient := &mocks.BlobstoreClient{}
	blobDownloader := NewHistoryBlobDownloader(blobstoreClient)
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		NextPageToken: []byte{1},
	})
	h.Error(err)
	h.Nil(resp)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Failed_CouldNotGetHighestVersion() {
	h.blobstoreClient.On("GetTags", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed to get tags")).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket: testArchivalBucket,
		DomainId:       testDomainID,
		WorkflowId:     testWorkflowID,
		RunId:          testRunID,
	})
	h.Error(err)
	h.Nil(resp)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Failed_CouldNotDownloadBlob() {
	h.blobstoreClient.On("GetTags", mock.Anything, mock.Anything, mock.Anything).Return(map[string]string{"1": ""}, nil).Once()
	h.blobstoreClient.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed to download blob")).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket: testArchivalBucket,
		DomainId:       testDomainID,
		WorkflowId:     testWorkflowID,
		RunId:          testRunID,
	})
	h.Error(err)
	h.Nil(resp)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Success_GetFirstPage() {
	unwrappedBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(common.FirstBlobPageToken),
			NextPageToken:    common.IntPtr(common.FirstBlobPageToken + 1),
			IsLast:           common.BoolPtr(false),
		},
		Body: &shared.History{},
	}
	bytes, err := json.Marshal(unwrappedBlob)
	h.NoError(err)
	historyBlob, err := blob.Wrap(blob.NewBlob(bytes, map[string]string{}), blob.JSONEncoded())
	h.blobstoreClient.On("GetTags", mock.Anything, mock.Anything, mock.Anything).Return(map[string]string{"1": "", "2": "", "10": ""}, nil).Once()
	historyKey, _ := NewHistoryBlobKey(testDomainID, testWorkflowID, testRunID, 10, common.FirstBlobPageToken)
	h.blobstoreClient.On("Download", mock.Anything, mock.Anything, historyKey).Return(historyBlob, nil).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket: testArchivalBucket,
		DomainId:       testDomainID,
		WorkflowId:     testWorkflowID,
		RunId:          testRunID,
	})
	h.NoError(err)
	h.Equal(hash(*unwrappedBlob), hash(*resp.HistoryBlob))
	expectedPageToken, _ := serializeHistoryTokenArchival(&getHistoryContinuationTokenArchival{
		BlobstorePageToken:   common.FirstBlobPageToken + 1,
		CloseFailoverVersion: 10,
	})
	h.Equal(expectedPageToken, resp.NextPageToken)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Success_ProvidedVersion() {
	unwrappedBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(common.FirstBlobPageToken),
			NextPageToken:    common.IntPtr(common.FirstBlobPageToken + 1),
			IsLast:           common.BoolPtr(false),
		},
		Body: &shared.History{},
	}
	bytes, err := json.Marshal(unwrappedBlob)
	h.NoError(err)
	historyBlob, err := blob.Wrap(blob.NewBlob(bytes, map[string]string{}), blob.JSONEncoded())
	historyKey, _ := NewHistoryBlobKey(testDomainID, testWorkflowID, testRunID, 10, common.FirstBlobPageToken)
	h.blobstoreClient.On("Download", mock.Anything, mock.Anything, historyKey).Return(historyBlob, nil).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket:       testArchivalBucket,
		DomainId:             testDomainID,
		WorkflowId:           testWorkflowID,
		RunId:                testRunID,
		CloseFailoverVersion: common.Int64Ptr(10),
	})
	h.NoError(err)
	h.Equal(hash(*unwrappedBlob), hash(*resp.HistoryBlob))
	expectedPageToken, _ := serializeHistoryTokenArchival(&getHistoryContinuationTokenArchival{
		BlobstorePageToken:   common.FirstBlobPageToken + 1,
		CloseFailoverVersion: 10,
	})
	h.Equal(expectedPageToken, resp.NextPageToken)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Success_UsePageToken() {
	unwrappedBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(common.FirstBlobPageToken + 1),
			NextPageToken:    common.IntPtr(common.FirstBlobPageToken + 2),
			IsLast:           common.BoolPtr(false),
		},
		Body: &shared.History{},
	}
	bytes, err := json.Marshal(unwrappedBlob)
	h.NoError(err)
	historyBlob, err := blob.Wrap(blob.NewBlob(bytes, map[string]string{}), blob.JSONEncoded())
	historyKey, _ := NewHistoryBlobKey(testDomainID, testWorkflowID, testRunID, 10, common.FirstBlobPageToken+1)
	h.blobstoreClient.On("Download", mock.Anything, mock.Anything, historyKey).Return(historyBlob, nil).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	requestPageToken, _ := serializeHistoryTokenArchival(&getHistoryContinuationTokenArchival{
		BlobstorePageToken:   common.FirstBlobPageToken + 1,
		CloseFailoverVersion: 10,
	})
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket: testArchivalBucket,
		DomainId:       testDomainID,
		WorkflowId:     testWorkflowID,
		RunId:          testRunID,
		NextPageToken:  requestPageToken,
	})
	h.NoError(err)
	h.Equal(hash(*unwrappedBlob), hash(*resp.HistoryBlob))
	expectedPageToken, _ := serializeHistoryTokenArchival(&getHistoryContinuationTokenArchival{
		BlobstorePageToken:   common.FirstBlobPageToken + 2,
		CloseFailoverVersion: 10,
	})
	h.Equal(expectedPageToken, resp.NextPageToken)
}

func (h *historyBlobDownloaderSuite) TestDownloadBlob_Success_GetLastPage() {
	unwrappedBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(common.FirstBlobPageToken + 1),
			NextPageToken:    common.IntPtr(common.FirstBlobPageToken + 2),
			IsLast:           common.BoolPtr(true),
		},
		Body: &shared.History{},
	}
	bytes, err := json.Marshal(unwrappedBlob)
	h.NoError(err)
	historyBlob, err := blob.Wrap(blob.NewBlob(bytes, map[string]string{}), blob.JSONEncoded())
	historyKey, _ := NewHistoryBlobKey(testDomainID, testWorkflowID, testRunID, 10, common.FirstBlobPageToken+1)
	h.blobstoreClient.On("Download", mock.Anything, mock.Anything, historyKey).Return(historyBlob, nil).Once()
	blobDownloader := NewHistoryBlobDownloader(h.blobstoreClient)
	requestPageToken, _ := serializeHistoryTokenArchival(&getHistoryContinuationTokenArchival{
		BlobstorePageToken:   common.FirstBlobPageToken + 1,
		CloseFailoverVersion: 10,
	})
	resp, err := blobDownloader.DownloadBlob(context.Background(), &DownloadPageRequest{
		ArchivalBucket: testArchivalBucket,
		DomainId:       testDomainID,
		WorkflowId:     testWorkflowID,
		RunId:          testRunID,
		NextPageToken:  requestPageToken,
	})
	h.NoError(err)
	h.Equal(hash(*unwrappedBlob), hash(*resp.HistoryBlob))
	h.Nil(resp.NextPageToken)
}
