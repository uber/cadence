package pinot

import (
	p "github.com/uber/cadence/common/persistence"
	"time"
)

type (
	// GenericClient is a generic interface for all versions of ElasticSearch clients
	GenericClient interface {
		// Search API is only for supporting various List[Open/Closed]WorkflowExecutions(ByXyz).
		// Use SearchByQuery or ScanByQuery for generic purpose searching.
		Search(request *SearchRequest) (*SearchResponse, error)
		// CountByQuery is for returning the count of workflow executions that match the query
		CountByQuery(query string) (int64, error)
	}

	// IsRecordValidFilter is a function to filter visibility records
	IsRecordValidFilter func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool

	// SearchRequest is request for Search
	SearchRequest struct {
		Query           string
		IsOpen          bool
		Filter          IsRecordValidFilter
		MaxResultWindow int
	}

	// GenericMatch is a match struct
	GenericMatch struct {
		Name string
		Text interface{}
	}

	// SearchByQueryRequest is request for SearchByQuery
	SearchByQueryRequest struct {
		Query           string
		NextPageToken   []byte
		PageSize        int
		Filter          IsRecordValidFilter
		MaxResultWindow int
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

	// GenericBackoff allows callers to implement their own Backoff strategy.
	GenericBackoff interface {
		// Next implements a BackoffFunc.
		Next(retry int) (time.Duration, bool)
	}
)
