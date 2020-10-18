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
	"math"
	"math/rand"
	"time"

	"github.com/uber/cadence/common/metrics"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"
)

// NewClient create a ES client
func NewGenericElasticSearchClient(
	connectConfig *config.ElasticSearchConfig,
	visibilityConfig *config.VisibilityConfig,
	logger log.Logger,
) (GenericElasticSearch, error) {
	// TODO hardcoded to V6 for now
	return newV6Client(connectConfig, visibilityConfig, logger)
}

type (
	// GenericElasticSearch is a generic interface for all versions of ElasticSearch clients
	GenericElasticSearch interface {
		Search(ctx context.Context, request *SearchRequest) (*p.InternalListWorkflowExecutionsResponse, error)
		SearchByQuery(ctx context.Context, request *SearchByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error)
		ScanByQuery(ctx context.Context, request *ScanByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error)
		CountByQuery(ctx context.Context, index, query string) (int64, error)
		PutMapping(ctx context.Context, index, root, key, valueType string) error
		CreateIndex(ctx context.Context, index string) error
		GetClosedWorkflowExecution(ctx context.Context, index string, request *p.InternalGetClosedWorkflowExecutionRequest) (*p.InternalGetClosedWorkflowExecutionResponse, error)

		RunBulkProcessor(ctx context.Context, p *GenericBulkProcessorParameters) (GenericBulkProcessor, error)
	}

	// SearchRequest is request for Search
	SearchRequest struct {
		Index       string
		ListRequest *p.InternalListWorkflowExecutionsRequest
		IsOpen      bool
		Filter      IsRecordValidFilter
		MatchQuery  *GenericMatch
	}

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
		Add(request *GenericBulkableAddRequest)
		Flush() error
		RetrieveKafkaKey(request GenericBulkableRequest, logger log.Logger, client metrics.Client) string
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
	GenericBulkAfterFunc func(executionId int64, requests []GenericBulkableRequest, response *GenericBulkResponse, err *GenericError)

	IsRecordValidFilter func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool

	// BulkableRequest is a generic interface to bulkable requests.

	GenericBulkableRequest interface {
		fmt.Stringer
		Source() ([]string, error)
	}

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

// ExponentialBackoff implements the simple exponential backoff described by
// Douglas Thain at http://dthain.blogspot.de/2009/02/exponential-backoff-in-distributed.html.
type ExponentialBackoff struct {
	t float64 // initial timeout (in msec)
	f float64 // exponential factor (e.g. 2)
	m float64 // maximum timeout (in msec)
}

// NewExponentialBackoff returns a ExponentialBackoff backoff policy.
// Use initialTimeout to set the first/minimal interval
// and maxTimeout to set the maximum wait interval.
func NewExponentialBackoff(initialTimeout, maxTimeout time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		t: float64(int64(initialTimeout / time.Millisecond)),
		f: 2.0,
		m: float64(int64(maxTimeout / time.Millisecond)),
	}
}

// Next implements BackoffFunc for ExponentialBackoff.
func (b *ExponentialBackoff) Next(retry int) (time.Duration, bool) {
	r := 1.0 + rand.Float64() // random number in [1..2]
	m := math.Min(r*b.t*math.Pow(b.f, float64(retry)), b.m)
	if m >= b.m {
		return 0, false
	}
	d := time.Duration(int64(m)) * time.Millisecond
	return d, true
}
