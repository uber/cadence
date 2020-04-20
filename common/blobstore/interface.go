package blobstore

import (
	"context"
)

type (
	// Client defines the interface to a blobstore client.
	Client interface {
		Put(context.Context, *PutRequest) (*PutResponse, error)
		Get(context.Context, *GetRequest) (*GetResponse, error)
		Exists(context.Context, *ExistsRequest) (*ExistsResponse, error)
		Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	}

	// PutRequest is the request to Put
	PutRequest struct {
		Key  string
		Blob Blob
	}

	// PutResponse is the response from Put
	PutResponse struct{}

	// GetRequest is the request to Get
	GetRequest struct {
		Key string
	}

	// GetResponse is the response from Get
	GetResponse struct {
		Blob Blob
	}

	// ExistsRequest is the request to Exists
	ExistsRequest struct {
		Key string
	}

	// ExistsResponse is the response from Exists
	ExistsResponse struct {
		Exists bool
	}

	// DeleteRequest is the request to Delete
	DeleteRequest struct {
		Key string
	}

	// DeleteResponse is the response from Delete
	DeleteResponse struct{}

	// Blob defines a blob which can be stored and fetched from blobstore
	Blob struct {
		Tags map[string]string
		Body []byte
	}
)
