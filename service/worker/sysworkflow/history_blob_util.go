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
		DomainName     *string `json:"domain_name"`
		DomainID       *string `json:"domain_id"`
		WorkflowID     *string `json:"workflow_id"`
		RunID          *string `json:"run_id"`
		PageToken      *string `json:"page_token"`
		Version        *string `json:"version"`
		UploadDateTime *string `json:"upload_date_time"`
		UploadCluster  *string `json:"update_cluster"`
		EventCount     *int64  `json:"event_count"`
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
