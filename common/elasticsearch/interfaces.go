// Copyright (c) 2020 Uber Technologies, Inc.
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

package elasticsearch

import (
	"context"
	"fmt"
	p "github.com/uber/cadence/common/persistence"
	"time"
)

type (
	// GenericElasticSearch is a generic interface for all versions of ElasticSearch clients
	GenericElasticSearch interface {
		Search(ctx context.Context, p *GenericSearchParameters) (*p.InternalListWorkflowExecutionsResponse, error)
		SearchByQuery(ctx context.Context, index, query string) (*p.InternalListWorkflowExecutionsResponse, error)
		ScrollFirstPage(ctx context.Context, index, query string, nextPageToken []byte) (*p.InternalListWorkflowExecutionsResponse, error)
		CountByQuery(ctx context.Context, index, query string) (int64, error)
		PutMapping(ctx context.Context, index, root, key, valueType string) error
		CreateIndex(ctx context.Context, index string) error

		RunBulkProcessor(ctx context.Context, p *GenericBulkProcessorParameters) (GenericBulkProcessor, error)
	}

	// GenericSearchParameters holds all required and optional parameters for executing a search
	GenericSearchParameters struct {
		Index       string
		Query       GenericQuery
		From        int
		PageSize    int
		Sorter      []GenericSorter
		SearchAfter []interface{}
	}

	GenericBulkProcessor interface {
		Start(ctx context.Context) error
		Stop() error
		Close() error
		Add(request GenericBulkableRequest)
		Flush() error
	}

	// GenericBulkProcessorParameters holds all required and optional parameters for executing bulk service
	GenericBulkProcessorParameters struct {
		Name          string
		NumOfWorkers  int
		BulkActions   int
		BulkSize      int
		FlushInterval time.Duration
		Backoff       GenericBackoff
		BeforeFunc    GenericBulkBeforeFunc
		AfterFunc     GenericBulkAfterFunc
	}

	// GenericSorter is an interface for sorting strategies
	GenericSorter interface {
		Source() (interface{}, error)
	}

	// GenericQuery represents the generic query interface.
	GenericQuery interface {
		// Source returns the JSON-serializable query request.
		Source() (interface{}, error)
	}

	// GenericBackoff allows callers to implement their own Backoff strategy.
	GenericBackoff interface {
		// Next implements a BackoffFunc.
		Next(retry int) (time.Duration, bool)
	}

	// GenericBulkBeforeFunc defines the signature of callbacks that are executed
	// before a commit to Elasticsearch.
	GenericBulkBeforeFunc func(executionId int64, requests []GenericBulkableRequest)

	// GenericBulkAfterFunc defines the signature of callbacks that are executed
	// after a commit to Elasticsearch. The err parameter signals an error.
	GenericBulkAfterFunc func(executionId int64, requests []GenericBulkableRequest, response *GenericBulkResponse, err error)

	// BulkableRequest is a generic interface to bulkable requests.
	GenericBulkableRequest interface {
		fmt.Stringer
		Source() ([]string, error)
	}

	// GenericBulkResponse is generic struct of bulk response
	GenericBulkResponse struct {
		Took   int                                   `json:"took,omitempty"`
		Errors bool                                  `json:"errors,omitempty"`
		Items  []map[string]*GenericBulkResponseItem `json:"items,omitempty"`
	}

	// GenericBulkResponseItem is the result of a single bulk request.
	GenericBulkResponseItem struct {
		Index         string `json:"_index,omitempty"`
		Type          string `json:"_type,omitempty"`
		Id            string `json:"_id,omitempty"`
		Version       int64  `json:"_version,omitempty"`
		Result        string `json:"result,omitempty"`
		SeqNo         int64  `json:"_seq_no,omitempty"`
		PrimaryTerm   int64  `json:"_primary_term,omitempty"`
		Status        int    `json:"status,omitempty"`
		ForcedRefresh bool   `json:"forced_refresh,omitempty"`
	}
)
