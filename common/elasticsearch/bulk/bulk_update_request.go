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

var _ GenericBulkableRequest = (*BulkUpdateRequest)(nil)

// BulkUpdateRequest is a request to update a document in Elasticsearch.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-bulk.html
// for details.
type BulkUpdateRequest struct {
	index string
	id    string

	version     int64
	versionType string // default is "internal"
	doc         interface{}

	source []string
}

type bulkUpdateRequestCommand map[string]bulkUpdateRequestCommandOp

type bulkUpdateRequestCommandOp struct {
	Index       string `json:"_index,omitempty"`
	ID          string `json:"_id,omitempty"`
	Version     int64  `json:"version,omitempty"`
	VersionType string `json:"version_type,omitempty"`
}

type bulkUpdateRequestCommandData struct {
	Doc    interface{} `json:"doc,omitempty"`
	Upsert interface{} `json:"upsert,omitempty"`
	Source *bool       `json:"_source,omitempty"`
}

// NewBulkUpdateRequest returns a new BulkUpdateRequest.
func NewBulkUpdateRequest() *BulkUpdateRequest {
	return &BulkUpdateRequest{}
}

// Index specifies the Elasticsearch index to use for this update request.
// If unspecified, the index set on the BulkService will be used.
func (r *BulkUpdateRequest) Index(index string) *BulkUpdateRequest {
	r.index = index
	r.source = nil
	return r
}

// ID specifies the identifier of the document to update.
func (r *BulkUpdateRequest) ID(id string) *BulkUpdateRequest {
	r.id = id
	r.source = nil
	return r
}

// Version indicates the version of the document as part of an optimistic
// concurrency model.
func (r *BulkUpdateRequest) Version(version int64) *BulkUpdateRequest {
	r.version = version
	r.source = nil
	return r
}

// VersionType can be "internal" (default), "external", "external_gte",
// or "external_gt".
func (r *BulkUpdateRequest) VersionType(versionType string) *BulkUpdateRequest {
	r.versionType = versionType
	r.source = nil
	return r
}

// Doc specifies the updated document.
func (r *BulkUpdateRequest) Doc(doc interface{}) *BulkUpdateRequest {
	r.doc = doc
	r.source = nil
	return r
}

// String returns the on-wire representation of the update request,
// concatenated as a single string.
func (r *BulkUpdateRequest) String() string {
	lines, err := r.Source()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return strings.Join(lines, "\n")
}

// Source returns the on-wire representation of the update request,
// split into an action-and-meta-data line and an (optional) source line.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-bulk.html
// for details.
func (r *BulkUpdateRequest) Source() ([]string, error) {
	// { "update" : { "_index" : "test", "_type" : "type1", "_id" : "1", ... } }
	// { "doc" : { "field1" : "value1", ... } }
	// or
	// { "update" : { "_index" : "test", "_type" : "type1", "_id" : "1", ... } }
	// { "script" : { ... } }

	if r.source != nil {
		return r.source, nil
	}

	lines := make([]string, 2)

	// "update" ...
	updateCommand := bulkUpdateRequestCommandOp{
		Index:       r.index,
		ID:          r.id,
		Version:     r.version,
		VersionType: r.versionType,
	}
	command := bulkUpdateRequestCommand{
		"update": updateCommand,
	}

	var err error
	var body []byte

	body, err = json.Marshal(command)

	if err != nil {
		return nil, err
	}

	lines[0] = string(body)

	// 2nd line: {"doc" : { ... }} or {"script": {...}}
	data := bulkUpdateRequestCommandData{
		Doc: r.doc,
	}

	// encoding/json
	body, err = json.Marshal(data)

	if err != nil {
		return nil, err
	}

	lines[1] = string(body)

	r.source = lines
	return lines, nil
}
