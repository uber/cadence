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

package filestore

import (
	"context"
	"errors"
	"os"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/service/config"
)

var (
	errInvalidFileMode = errors.New("invalid file mode")
	errInvalidDirMode  = errors.New("invalid directory mode")
)

var (
	errDirectoryExpected  = errors.New("a path to a directory was expected")
	errFileExpected       = errors.New("a path to a file was expected")
	errEmptyDirectoryPath = errors.New("directory path is empty")
)

type (
	client struct {
		outputDirectory string
	}
)

func NewFileBlobstore(cfg *config.FileBlobstore) (blobstore.Client, error) {
	if cfg == nil || len(cfg.OutputDirectory) {
		// do some other verifications here or whatever
	}
	return &client{
		outputDirectory: cfg.OutputDirectory,
	}, nil
}

func (c *client) Put(_ context.Context, request *blobstore.PutRequest) (*blobstore.PutResponse, error) {
	return nil, nil
}

func (c *client) Get(context.Context, *blobstore.GetRequest) (*blobstore.GetResponse, error) {
	return nil, nil
}

func (c *client) Exists(context.Context, *blobstore.ExistsRequest) (*blobstore.ExistsResponse, error) {
	return nil, nil
}

func (c *client) Delete(context.Context, *blobstore.DeleteRequest) (*blobstore.DeleteResponse, error) {
	return nil, nil
}

func validateDirPath(dirPath string) error {
	if len(dirPath) == 0 {
		return errEmptyDirectoryPath
	}
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errDirectoryExpected
	}
	return nil
}
