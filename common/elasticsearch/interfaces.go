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
	"time"

	"github.com/uber/cadence/common/metrics"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"
)

// NewGenericClient create a ES client
func NewGenericClient(
	connectConfig *config.ElasticSearchConfig,
	visibilityConfig *config.VisibilityConfig,
	logger log.Logger,
) (GenericClient, error) {
	// TODO hardcoded to V6 for now
	return newV6Client(connectConfig, visibilityConfig, logger)
}

type (
	// GenericClient is a generic interface for all versions of ElasticSearch clients
	GenericClient interface {
		// Search API is only for supporting various List[Open/Closed]WorkflowExecutions(ByXyz).
		// Use SearchByQuery or ScanByQuery for generic purpose searching.
		Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error)
		// SearchByQuery is the generic purpose searching
		SearchByQuery(ctx context.Context, request *SearchByQueryRequest) (*SearchResponse, error)
		// ScanByQuery is also generic purpose searching, but implemented with ScrollService of ElasticSearch,
		// which is more performant for pagination, but comes with some limitation of in-parallel requests.
		ScanByQuery(ctx context.Context, request *ScanByQueryRequest) (*SearchResponse, error)
		// TODO remove it in https://github.com/uber/cadence/issues/3682
		SearchForOneClosedExecution(ctx context.Context, index string, request *SearchForOneClosedExecutionRequest) (*SearchForOneClosedExecutionResponse, error)
		// CountByQuery is for returning the count of workflow executions that match the query
		CountByQuery(ctx context.Context, index, query string) (int64, error)

		// RunBulkProcessor returns a processor for adding/removing docs into ElasticSearch index
		RunBulkProcessor(ctx context.Context, p *BulkProcessorParameters) (GenericBulkProcessor, error)

		// PutMapping adds new field type to the index
		PutMapping(ctx context.Context, index, root, key, valueType string) error
		// CreateIndex creates a new index
		CreateIndex(ctx context.Context, index string) error
	}

	// SearchRequest is request for Search
	SearchRequest struct {
		Index       string
		ListRequest *p.InternalListWorkflowExecutionsRequest
		IsOpen      bool
		Filter      IsRecordValidFilter
		MatchQuery  *GenericMatch
	}

	// GenericMatch is a match struct
	GenericMatch struct {
		Name string
		Text interface{}
	}

	// SearchByQueryRequest is request for SearchByQuery
	SearchByQueryRequest struct {
		Index         string
		Query         string
		NextPageToken []byte
		PageSize      int
		Filter        IsRecordValidFilter
	}

	// ScanByQueryRequest is request for SearchByQuery
	ScanByQueryRequest struct {
		Index         string
		Query         string
		NextPageToken []byte
		PageSize      int
	}

	// SearchResponse is a response to Search, SearchByQuery and ScanByQuery
	SearchResponse = p.InternalListWorkflowExecutionsResponse

	// SearchForOneClosedExecutionRequest is request for SearchForOneClosedExecution
	SearchForOneClosedExecutionRequest = p.InternalGetClosedWorkflowExecutionRequest

	// SearchForOneClosedExecutionResponse is response for SearchForOneClosedExecution
	SearchForOneClosedExecutionResponse = p.InternalGetClosedWorkflowExecutionResponse

	// GenericBulkProcessor is a bulk processor
	GenericBulkProcessor interface {
		Start(ctx context.Context) error
		Stop() error
		Close() error
		Add(request *GenericBulkableAddRequest)
		Flush() error
		RetrieveKafkaKey(request GenericBulkableRequest, logger log.Logger, client metrics.Client) string
	}

	// BulkProcessorParameters holds all required and optional parameters for executing bulk service
	BulkProcessorParameters struct {
		Name          string
		NumOfWorkers  int
		BulkActions   int
		BulkSize      int
		FlushInterval time.Duration
		Backoff       GenericBackoff
		BeforeFunc    GenericBulkBeforeFunc
		AfterFunc     GenericBulkAfterFunc
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
	GenericBulkAfterFunc func(executionId int64, requests []GenericBulkableRequest, response *GenericBulkResponse, err *GenericError)

	// IsRecordValidFilter is a function to filter visibility records
	IsRecordValidFilter func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool

	// GenericBulkableRequest is a generic interface to bulkable requests.
	GenericBulkableRequest interface {
		fmt.Stringer
		Source() ([]string, error)
	}

	// GenericBulkableAddRequest a struct to hold a bulk request
	GenericBulkableAddRequest struct {
		Index       string
		Type        string
		Id          string
		VersionType string
		Version     int64
		// true means it's delete, otherwise it's a index request
		IsDelete bool
		// should be nil if IsDelete is true
		Doc interface{}
	}

	// GenericBulkResponse is generic struct of bulk response
	GenericBulkResponse struct {
		Took   int                                   `json:"took,omitempty"`
		Errors bool                                  `json:"errors,omitempty"`
		Items  []map[string]*GenericBulkResponseItem `json:"items,omitempty"`
	}

	// GenericError encapsulates error status and details returned from Elasticsearch.
	GenericError struct {
		Status  int   `json:"status"`
		Details error `json:"error,omitempty"`
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
		// the error details
		Error interface{}
	}

	// VisibilityRecord is a struct of doc for deserialization
	VisibilityRecord struct {
		WorkflowID    string
		RunID         string
		WorkflowType  string
		StartTime     int64
		ExecutionTime int64
		CloseTime     int64
		CloseStatus   workflow.WorkflowExecutionCloseStatus
		HistoryLength int64
		Memo          []byte
		Encoding      string
		TaskList      string
		Attr          map[string]interface{}
	}
)
