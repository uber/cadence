// Copyright (c) 2021 Uber Technologies, Inc.
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

package cadence

// below are the names of all mongoDB collections
const (
	ClusterConfigCollectionName = "cluster_config"
)

// NOTE1: MongoDB collection is schemaless -- there is no schema file for collection. We use Go lang structs to define the collection fields.

// NOTE2: MongoDB doesn't allow using camel case or underscore in the field names

// ClusterConfigCollectionEntry is the schema of configStore
// IMPORTANT: making change to this struct is changing the MongoDB collection schema. Please make sure it's backward compatible(e.g., don't delete the field, or change the annotation value).
type ClusterConfigCollectionEntry struct {
	ID                   int    `json:"_id,omitempty"`
	RowType              int    `json:"rowtype"`
	Version              int64  `json:"version"`
	Data                 []byte `json:"data"`
	DataEncoding         string `json:"dataencoding"`
	UnixTimestampSeconds int64  `json:"unixtimestampseconds"`
}
