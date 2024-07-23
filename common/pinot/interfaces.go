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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination generic_client_mock.go -self_package github.com/uber/cadence/common/pinot

package pinot

import p "github.com/uber/cadence/common/persistence"

type (
	// GenericClient is a generic interface for all versions of Pinot clients
	GenericClient interface {
		// Search API is only for supporting various List[Open/Closed]WorkflowExecutions(ByXyz).
		// Use SearchByQuery or ScanByQuery for generic purpose searching.
		Search(request *SearchRequest) (*SearchResponse, error)
		SearchAggr(request *SearchRequest) (AggrResponse, error)
		// CountByQuery is for returning the count of workflow executions that match the query
		CountByQuery(query string) (int64, error)
		GetTableName() string
	}

	// IsRecordValidFilter is a function to filter visibility records
	IsRecordValidFilter func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool

	// SearchRequest is request for Search
	SearchRequest struct {
		Query           string
		IsOpen          bool
		Filter          IsRecordValidFilter
		MaxResultWindow int
		ListRequest     *p.InternalListWorkflowExecutionsRequest
	}

	// SearchResponse is a response to Search, SearchByQuery and ScanByQuery
	SearchResponse = p.InternalListWorkflowExecutionsResponse
	AggrResponse   [][]interface{}
)
