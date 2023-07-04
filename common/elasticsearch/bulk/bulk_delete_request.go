// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package bulk

import (
	"encoding/json"
	"fmt"
	"strings"
)

var _ GenericBulkableRequest = (*BulkDeleteRequest)(nil)

// BulkDeleteRequest is a request to remove a document from Elasticsearch.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-bulk.html
// for details.
type BulkDeleteRequest struct {
	index       string
	typ         string
	id          string
	routing     string
	version     int64
	versionType string // default is "internal"

	source []string
}

type bulkDeleteRequestCommand map[string]bulkDeleteRequestCommandOp

type bulkDeleteRequestCommandOp struct {
	Index       string `json:"_index,omitempty"`
	ID          string `json:"_id,omitempty"`
	Version     int64  `json:"version,omitempty"`
	VersionType string `json:"version_type,omitempty"`
}

// NewBulkDeleteRequest returns a new BulkDeleteRequest.
func NewBulkDeleteRequest() *BulkDeleteRequest {
	return &BulkDeleteRequest{}
}

// Index specifies the Elasticsearch index to use for this delete request.
// If unspecified, the index set on the BulkService will be used.
func (r *BulkDeleteRequest) Index(index string) *BulkDeleteRequest {
	r.index = index
	r.source = nil
	return r
}

// Type specifies the Elasticsearch type to use for this delete request.
// If unspecified, the type set on the BulkService will be used.
func (r *BulkDeleteRequest) Type(typ string) *BulkDeleteRequest {
	r.typ = typ
	r.source = nil
	return r
}

// ID specifies the identifier of the document to delete.
func (r *BulkDeleteRequest) ID(id string) *BulkDeleteRequest {
	r.id = id
	r.source = nil
	return r
}

// Version indicates the version to be deleted as part of an optimistic
// concurrency model.
func (r *BulkDeleteRequest) Version(version int64) *BulkDeleteRequest {
	r.version = version
	r.source = nil
	return r
}

// VersionType can be "internal" (default), "external", "external_gte",
// or "external_gt".
func (r *BulkDeleteRequest) VersionType(versionType string) *BulkDeleteRequest {
	r.versionType = versionType
	r.source = nil
	return r
}

// String returns the on-wire representation of the delete request,
// concatenated as a single string.
func (r *BulkDeleteRequest) String() string {
	lines, err := r.Source()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return strings.Join(lines, "\n")
}

// Source returns the on-wire representation of the delete request,
// split into an action-and-meta-data line and an (optional) source line.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-bulk.html
// for details.
func (r *BulkDeleteRequest) Source() ([]string, error) {
	if r.source != nil {
		return r.source, nil
	}
	command := bulkDeleteRequestCommand{
		"delete": bulkDeleteRequestCommandOp{
			Index:       r.index,
			ID:          r.id,
			Version:     r.version,
			VersionType: r.versionType,
		},
	}

	var err error
	var body []byte

	body, err = json.Marshal(command)

	if err != nil {
		return nil, err
	}

	lines := []string{string(body)}
	r.source = lines

	return lines, nil
}
