package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/service/config"
	"os"
)

type (
	client struct {
		outputDirectory string
	}
)

// NewFilestoreClient constructs a blobstore backed by local file system
func NewFilestoreClient(cfg *config.FileBlobstore) (blobstore.Client, error) {
	if cfg == nil {
		return nil, errors.New("file blobstore config is nil")
	}
	if len(cfg.OutputDirectory) == 0 {
		return nil, errors.New("output directory not given for file blobstore")
	}
	outputDirectory := cfg.OutputDirectory
	exists, err := common.DirectoryExists(outputDirectory)
	if err != nil {
		return nil, err
	}
	if exists {
		return &client{
			outputDirectory: outputDirectory,
		}, nil
	}
	if err := common.MkdirAll(outputDirectory, os.FileMode(0766)); err != nil {
		return nil, err
	}
	return &client{
		outputDirectory: outputDirectory,
	}, nil
}

// Put stores a blob
func (c *client) Put(_ context.Context, request *blobstore.PutRequest) (*blobstore.PutResponse, error) {
	data, err := c.serializeBlob(request.Blob)
	if err != nil {
		return nil, err
	}
	if err := common.WriteFile(c.filepath(request.Key), data, os.FileMode(0666)); err != nil {
		return nil, err
	}
	return &blobstore.PutResponse{}, nil
}

// Get fetches a blob
func (c *client) Get(_ context.Context, request *blobstore.GetRequest) (*blobstore.GetResponse, error) {
	data, err := common.ReadFile(c.filepath(request.Key))
	if err != nil {
		return nil, err
	}
	blob, err := c.deserializeBlob(data)
	if err != nil {
		return nil, err
	}
	return &blobstore.GetResponse{
		Blob: *blob,
	}, nil
}

// Exists determines if a blob exists
func (c *client) Exists(_ context.Context, request *blobstore.ExistsRequest) (*blobstore.ExistsResponse, error) {
	exists, err := common.FileExists(c.filepath(request.Key))
	if err != nil {
		return nil, err
	}
	return &blobstore.ExistsResponse{
		Exists: exists,
	}, nil
}

// Delete deletes a blob
func (c *client) Delete(_ context.Context, request *blobstore.DeleteRequest) (*blobstore.DeleteResponse, error) {
	if err := os.Remove(c.filepath(request.Key)); err != nil {
		return nil, err
	}
	return &blobstore.DeleteResponse{}, nil
}

func (c *client) deserializeBlob(data []byte) (*blobstore.Blob, error) {
	var blob blobstore.Blob
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, err
	}
	return &blob, nil
}

func (c *client) serializeBlob(blob blobstore.Blob) ([]byte, error) {
	return json.Marshal(blob)
}

func (c *client) filepath(key string) string {
	return fmt.Sprintf("%v/%v", c.outputDirectory, key)
}