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

package sysworkflow

import (
	"code.uber.internal/devexp/cadence-server/.tmp/.go/goroot/src/encoding/json"
	"fmt"
	"github.com/dgryski/go-farm"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/blobstore"
	"strings"
)

const (
	// HistoryBlobKeyExt is the blob key extension on all history blobs
	HistoryBlobKeyExt = ".history"
)

type (
	// HistoryBlobHeader is the header attached to all history blobs
	HistoryBlobHeader struct {
		DomainName           *string `json:"domain_name"`
		DomainID             *string `json:"domain_id"`
		WorkflowID           *string `json:"workflow_id"`
		RunID                *string `json:"run_id"`
		PageToken            *string `json:"page_token"`
		StartFailoverVersion *string `json:"start_failover_version"`
		CloseFailoverVersion *string `json"close_failover_version"`
		StartEventID         *int64  `json:"start_event_id"`
		CloseEventID         *int64  `json:"close_event_id"`
		UploadDateTime       *string `json:"upload_date_time"`
		UploadCluster        *string `json:"upload_cluster"`
		EventCount           *int64  `json:"event_count"`
	}

	// HistoryBlob is the serialized that histories get uploaded as
	HistoryBlob struct {
		Header *HistoryBlobHeader `json:"header"`
		Events *shared.History    `json:"events"`
	}
)

// NewHistoryBlobKey returns a key for history blob
func NewHistoryBlobKey(domainID, workflowID, runID, pageToken, version string) (blobstore.Key, error) {
	hashInput := strings.Join([]string{domainID, workflowID, runID}, "")
	hash := fmt.Sprintf("%v", farm.Fingerprint64([]byte(hashInput)))
	return blobstore.NewKey(HistoryBlobKeyExt, hash, pageToken, version)
}

// TODO: come up with some better naming for this method
// perhaps there is not such method as this maybe there is just a method that given historyBlob marshall it
// given a history blob header construct the tags
// these two methods are easily testable
// they allow me to construct a blob with a single line and two method calls
func MakeBlobOrWhateverMethodName(historyBlob *HistoryBlob) (blobstore.Blob, error) {

	// TODO: this code all needs to be cleaned up, factored in different methods and tested
	body, err := json.Marshal(historyBlob)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	inrec, err := json.Marshal(historyBlob.Header)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(inrec, &m); err != nil {
		return nil, err
	}

	var tags map[string]string
	for k, v := range m {
		tags[k] = fmt.Sprintf("%v", v)
	}

	return blobstore.NewBlob(body, tags), nil
}

// maybe you make it symmetric so given a tag map[string]string you can get the header
// given a byte array which is the blob body you can construct the HistoryBlob again

// That way we have four small methods which are self contained and give us all the flexibility we need
