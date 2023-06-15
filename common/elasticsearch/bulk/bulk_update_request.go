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

// BulkUpdateRequest is a request to update a document in Elasticsearch.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-bulk.html
// for details.
type BulkUpdateRequest struct {
	GenericBulkableRequest
	index string
	typ   string
	id    string

	routing         string
	parent          string
	version         int64
	versionType     string // default is "internal"
	retryOnConflict *int
	upsert          interface{}
	docAsUpsert     *bool
	detectNoop      *bool
	doc             interface{}
	returnSource    *bool
	ifSeqNo         *int64
	ifPrimaryTerm   *int64

	source []string
}

type bulkUpdateRequestCommand map[string]bulkUpdateRequestCommandOp

type bulkUpdateRequestCommandOp struct {
	Index  string `json:"_index,omitempty"`
	Type   string `json:"_type,omitempty"`
	ID     string `json:"_id,omitempty"`
	Parent string `json:"parent,omitempty"`
	// RetryOnConflict is "_retry_on_conflict" for 6.0 and "retry_on_conflict" for 6.1+.
	RetryOnConflict *int   `json:"retry_on_conflict,omitempty"`
	Routing         string `json:"routing,omitempty"`
	Version         int64  `json:"version,omitempty"`
	VersionType     string `json:"version_type,omitempty"`
	IfSeqNo         *int64 `json:"if_seq_no,omitempty"`
	IfPrimaryTerm   *int64 `json:"if_primary_term,omitempty"`
}

type bulkUpdateRequestCommandData struct {
	DetectNoop     *bool       `json:"detect_noop,omitempty"`
	Doc            interface{} `json:"doc,omitempty"`
	DocAsUpsert    *bool       `json:"doc_as_upsert,omitempty"`
	Script         interface{} `json:"script,omitempty"`
	ScriptedUpsert *bool       `json:"scripted_upsert,omitempty"`
	Upsert         interface{} `json:"upsert,omitempty"`
	Source         *bool       `json:"_source,omitempty"`
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

// Type specifies the Elasticsearch type to use for this update request.
// If unspecified, the type set on the BulkService will be used.
func (r *BulkUpdateRequest) Type(typ string) *BulkUpdateRequest {
	r.typ = typ
	r.source = nil
	return r
}

// ID specifies the identifier of the document to update.
func (r *BulkUpdateRequest) ID(id string) *BulkUpdateRequest {
	r.id = id
	r.source = nil
	return r
}

// Routing specifies a routing value for the request.
func (r *BulkUpdateRequest) Routing(routing string) *BulkUpdateRequest {
	r.routing = routing
	r.source = nil
	return r
}

// Parent specifies the identifier of the parent document (if available).
func (r *BulkUpdateRequest) Parent(parent string) *BulkUpdateRequest {
	r.parent = parent
	r.source = nil
	return r
}

// RetryOnConflict specifies how often to retry in case of a version conflict.
func (r *BulkUpdateRequest) RetryOnConflict(retryOnConflict int) *BulkUpdateRequest {
	r.retryOnConflict = &retryOnConflict
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

// IfSeqNo indicates to only perform the index operation if the last
// operation that has changed the document has the specified sequence number.
func (r *BulkUpdateRequest) IfSeqNo(ifSeqNo int64) *BulkUpdateRequest {
	r.ifSeqNo = &ifSeqNo
	return r
}

// IfPrimaryTerm indicates to only perform the index operation if the
// last operation that has changed the document has the specified primary term.
func (r *BulkUpdateRequest) IfPrimaryTerm(ifPrimaryTerm int64) *BulkUpdateRequest {
	r.ifPrimaryTerm = &ifPrimaryTerm
	return r
}

// Doc specifies the updated document.
func (r *BulkUpdateRequest) Doc(doc interface{}) *BulkUpdateRequest {
	r.doc = doc
	r.source = nil
	return r
}

// DocAsUpsert indicates whether the contents of Doc should be used as
// the Upsert value.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/docs-update.html#_literal_doc_as_upsert_literal
// for details.
func (r *BulkUpdateRequest) DocAsUpsert(docAsUpsert bool) *BulkUpdateRequest {
	r.docAsUpsert = &docAsUpsert
	r.source = nil
	return r
}

// DetectNoop specifies whether changes that don't affect the document
// should be ignored (true) or unignored (false). This is enabled by default
// in Elasticsearch.
func (r *BulkUpdateRequest) DetectNoop(detectNoop bool) *BulkUpdateRequest {
	r.detectNoop = &detectNoop
	r.source = nil
	return r
}

// Upsert specifies the document to use for upserts. It will be used for
// create if the original document does not exist.
func (r *BulkUpdateRequest) Upsert(doc interface{}) *BulkUpdateRequest {
	r.upsert = doc
	r.source = nil
	return r
}

// ReturnSource specifies whether Elasticsearch should return the source
// after the update. In the request, this responds to the `_source` field.
// It is false by default.
func (r *BulkUpdateRequest) ReturnSource(source bool) *BulkUpdateRequest {
	r.returnSource = &source
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
		Index:           r.index,
		Type:            r.typ,
		ID:              r.id,
		Routing:         r.routing,
		Parent:          r.parent,
		Version:         r.version,
		VersionType:     r.versionType,
		RetryOnConflict: r.retryOnConflict,
		IfSeqNo:         r.ifSeqNo,
		IfPrimaryTerm:   r.ifPrimaryTerm,
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
	var doc interface{}
	if r.doc != nil {
		// Automatically serialize strings as raw JSON
		switch t := r.doc.(type) {
		default:
			doc = r.doc
		case string:
			if len(t) > 0 {
				doc = json.RawMessage(t)
			}
		case *string:
			if t != nil && len(*t) > 0 {
				doc = json.RawMessage(*t)
			}
		}
	}
	data := bulkUpdateRequestCommandData{
		DocAsUpsert: r.docAsUpsert,
		DetectNoop:  r.detectNoop,
		Upsert:      r.upsert,
		Doc:         doc,
		Source:      r.returnSource,
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
