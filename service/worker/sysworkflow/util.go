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
	"strings"
)

const (
	historyBlobFilenameExt = ".history"
)

// HistoryBlobFilename constructs name of history file from domainID, workflowID and runID
func HistoryBlobFilename(domainID string, workflowID string, runID string) string {
	hashInput := strings.Join([]string{domainID, workflowID, runID}, "")
	hash := farm.Fingerprint64([]byte(hashInput))
	return fmt.Sprintf("%v%v", hash, historyBlobFilenameExt)
}

// TODO: I should be constructing file names correctly
// TODO: Given user tags and history events I should be able to construct a JSON byte array


// file names should take as input the following
/**
- domainID, workflowID, runID (under this is different pages, under each page is different versions of that page)
- pageNumber
- historyVersion
 */
