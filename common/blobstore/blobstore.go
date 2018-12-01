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
	"context"
	"errors"
	"io"
	"time"
)

type BlobMetadata struct {
	CreatedAt *time.Time
	Owner     *string
	Size      *int64
	Tags      map[string]string
}

type Blob struct {
	Size *int64 // optional
	Body io.Reader
}

// ListItem is a single listing in ListByPrefixResponse
type ListItem struct {
	IsBlob       *bool
	Name         *string
	BlobMetadata BlobMetadata
}

// UploadBlobRequest is request for UploadBlob
type UploadBlobRequest struct {
	Path  *string
	Index *bool
	Blob  Blob
	Tags  map[string]string
}

// DownloadBlobRequest is request for DownloadBlob
type DownloadBlobRequest struct {
	Path *string
}

// GetBlobMetadataRequest is request for GetBlobMetadata
type GetBlobMetadataRequest struct {
	Path *string
}

// GetBlobMetadataResponse is response from GetBlobMetadata
type GetBlobMetadataResponse struct {
	BlobMetadata BlobMetadata
}

// ListByPrefixRequest is request for ListByPrefix
type ListByPrefixRequest struct {
	Path     *string
	ListFrom *string
	Detailed *bool
	Limit    *int
}

// ListByPrefixResponse is response from ListByPrefix
type ListByPrefixResponse struct {
	ListItems []*ListItem
}

// Client is the interface to interact with archival store
type Client interface {
	UploadBlob(ctx context.Context, request *UploadBlobRequest) error
	DownloadBlob(ctx context.Context, request *DownloadBlobRequest) (Blob, error)
	GetBlobMetadata(ctx context.Context, request *GetBlobMetadataRequest) (*GetBlobMetadataResponse, error)
	ListByPrefix(ctx context.Context, request *ListByPrefixRequest) (*ListByPrefixResponse, error)
}

type nopClient struct{}

func (mc *nopClient) UploadBlob(ctx context.Context, request *UploadBlobRequest) error {
	return errors.New("method not supported")
}

func (mc *nopClient) DownloadBlob(ctx context.Context, request *DownloadBlobRequest) (Blob, error) {
	return Blob{}, errors.New("method not supported")
}

func (mc *nopClient) GetBlobMetadata(ctx context.Context, request *GetBlobMetadataRequest) (*GetBlobMetadataResponse, error) {
	return nil, errors.New("method not supported")
}

func (mc *nopClient) ListByPrefix(ctx context.Context, request *ListByPrefixRequest) (*ListByPrefixResponse, error) {
	return nil, errors.New("method not supported")
}

// NewNopClient creates a nop client
func NewNopClient() Client {
	return &nopClient{}
}
