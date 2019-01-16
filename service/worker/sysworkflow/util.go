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
	historyBlobKeyExt = ".history"
)

func HistoryBlobKey(domainID, workflowID, runID, pageToken, version string) (blobstore.Key, error) {
	hashInput := strings.Join([]string{domainID, workflowID, runID}, "")
	hash := fmt.Sprintf("%v", farm.Fingerprint64([]byte(hashInput)))
	return blobstore.NewKey(historyBlobKeyExt, hash, pageToken, version)
}

type historyBlobHeader struct {
	domainName string `json:"domain_name"`
	domainID string `json:"domain_id"`
	workflowID string `json:"workflow_id"`
	runID string `json:"run_id"`
	pageToken string `json:"page_token"`
	version string `json:"version"`
	uploadDateTime string `json:"upload_date_time"`
	uploadCluster string `json:"update_cluster"`
	eventCount int64 `json:"event_count"`
}

type historyBlob struct {
	header *historyBlobHeader `json:"header"`
	events *shared.History `json:"events"`
}

func HistoryBlob() ([]byte, error) {

}