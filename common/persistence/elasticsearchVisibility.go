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

package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/logging"
)

const esPersistenceName = "elasticsearch"

type (
	esVisibilityManager struct {
		esClient *elastic.Client
		index    string
		logger   bark.Logger
	}

	esVisibilityPageToken struct {
		From int
	}

	visibilityRecord struct {
		WorkflowID    string
		RunID         string
		WorkflowType  string
		StartTime     int64
		CloseTime     int64
		CloseStatus   workflow.WorkflowExecutionCloseStatus
		HistoryLength int64
	}
)

var _ VisibilityManager = (*esVisibilityManager)(nil)

var (
	errNotSupported = errors.New("operation not support")
)

// NewElasticSearchVisibilityManager create a visibility manager connecting to ElasticSearch
func NewElasticSearchVisibilityManager(esClient *elastic.Client, index string, logger bark.Logger) (VisibilityManager, error) {

	return &esVisibilityManager{
		esClient: esClient,
		index:    index,
		logger:   logger.WithField(logging.TagWorkflowComponent, logging.TagValueESVisibilityManager),
	}, nil
}

func (v *esVisibilityManager) Close() {}

func (v *esVisibilityManager) GetName() string {
	return esPersistenceName
}

func (v *esVisibilityManager) RecordWorkflowExecutionStarted(request *RecordWorkflowExecutionStartedRequest) error {
	return errNotSupported
}

func (v *esVisibilityManager) RecordWorkflowExecutionClosed(request *RecordWorkflowExecutionClosedRequest) error {
	return errNotSupported
}

func (v *esVisibilityManager) ListOpenWorkflowExecutions(
	request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := true
	searchResult, err := v.getSearchResult(request, token, nil, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutions failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListClosedWorkflowExecutions(
	request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := false
	searchResult, err := v.getSearchResult(request, token, nil, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListOpenWorkflowExecutionsByType(
	request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := true
	matchQuery := elastic.NewMatchQuery(es.WorkflowType, request.WorkflowTypeName)
	searchResult, err := v.getSearchResult(&request.ListWorkflowExecutionsRequest, token, matchQuery, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutionsByType failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListClosedWorkflowExecutionsByType(
	request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := false
	matchQuery := elastic.NewMatchQuery(es.WorkflowType, request.WorkflowTypeName)
	searchResult, err := v.getSearchResult(&request.ListWorkflowExecutionsRequest, token, matchQuery, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByType failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListOpenWorkflowExecutionsByWorkflowID(
	request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := true
	matchQuery := elastic.NewMatchQuery(es.WorkflowID, request.WorkflowID)
	searchResult, err := v.getSearchResult(&request.ListWorkflowExecutionsRequest, token, matchQuery, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutionsByWorkflowID failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListClosedWorkflowExecutionsByWorkflowID(
	request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := false
	matchQuery := elastic.NewMatchQuery(es.WorkflowID, request.WorkflowID)
	searchResult, err := v.getSearchResult(&request.ListWorkflowExecutionsRequest, token, matchQuery, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByWorkflowID failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) ListClosedWorkflowExecutionsByStatus(
	request *ListClosedWorkflowExecutionsByStatusRequest) (*ListWorkflowExecutionsResponse, error) {

	token, err := v.getNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	isOpen := false
	matchQuery := elastic.NewMatchQuery(es.CloseStatus, int32(request.Status))
	searchResult, err := v.getSearchResult(&request.ListWorkflowExecutionsRequest, token, matchQuery, isOpen)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByStatus failed. Error: %v", err),
		}
	}

	return v.getListWorkflowExecutionsResponse(searchResult.Hits, token, isOpen)
}

func (v *esVisibilityManager) GetClosedWorkflowExecution(
	request *GetClosedWorkflowExecutionRequest) (*GetClosedWorkflowExecutionResponse, error) {
	return nil, errNotSupported
}

func (v *esVisibilityManager) getNextPageToken(token []byte) (*esVisibilityPageToken, error) {
	var result *esVisibilityPageToken
	var err error
	if len(token) > 0 {
		result, err = v.deserializePageToken(token)
		if err != nil {
			return nil, &workflow.BadRequestError{
				Message: fmt.Sprintf("not able to deserialize page token in request. err: %v", err),
			}
		}
	} else {
		result = &esVisibilityPageToken{From: 0}
	}
	return result, nil
}

func (v *esVisibilityManager) getSearchResult(request *ListWorkflowExecutionsRequest, token *esVisibilityPageToken,
	matchQuery *elastic.MatchQuery, isOpen bool) (*elastic.SearchResult, error) {

	matchDomainQuery := elastic.NewMatchQuery(es.DomainID, request.DomainUUID)
	existClosedStatusQuery := elastic.NewExistsQuery(es.CloseStatus)
	var rangeQuery *elastic.RangeQuery
	if isOpen {
		rangeQuery = elastic.NewRangeQuery(es.StartTime)
	} else {
		rangeQuery = elastic.NewRangeQuery(es.CloseTime)
	}
	rangeQuery = rangeQuery.Gte(request.EarliestStartTime).Lte(request.LatestStartTime) // TODO rename request fields for close time

	boolQuery := elastic.NewBoolQuery().Must(matchDomainQuery).Filter(rangeQuery)
	if matchQuery != nil {
		boolQuery = boolQuery.Must(matchQuery)
	}
	if isOpen {
		boolQuery = boolQuery.MustNot(existClosedStatusQuery)
	} else {
		boolQuery = boolQuery.Must(existClosedStatusQuery)
	}

	ctx := context.Background()
	search := v.esClient.Search().
		Index(v.index).
		Query(boolQuery).
		From(token.From).
		Size(request.PageSize)
	if isOpen {
		search = search.Sort(es.StartTime, false)
	} else {
		search = search.Sort(es.CloseTime, false)
	}
	return search.Do(ctx)
}

func (v *esVisibilityManager) getListWorkflowExecutionsResponse(searchHits *elastic.SearchHits,
	token *esVisibilityPageToken, isOpen bool) (*ListWorkflowExecutionsResponse, error) {

	response := &ListWorkflowExecutionsResponse{}
	actualHits := searchHits.Hits
	numOfActualHits := len(actualHits)

	nextPageToken, err := v.serializePageToken(&esVisibilityPageToken{From: token.From + numOfActualHits})
	if err != nil {
		return nil, &workflow.BadRequestError{Message: "unable to serialize page token"}
	}
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	response.Executions = make([]*workflow.WorkflowExecutionInfo, 0)
	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo := v.convertSearchResultToVisibilityRecord(actualHits[i], isOpen)
		response.Executions = append(response.Executions, workflowExecutionInfo)
	}

	return response, nil
}

func (v *esVisibilityManager) deserializePageToken(data []byte) (*esVisibilityPageToken, error) {
	var token esVisibilityPageToken
	err := json.Unmarshal(data, &token)
	return &token, err
}

func (v *esVisibilityManager) serializePageToken(token *esVisibilityPageToken) ([]byte, error) {
	data, err := json.Marshal(token)
	return data, err
}

func (v *esVisibilityManager) convertSearchResultToVisibilityRecord(hit *elastic.SearchHit, isOpen bool) *workflow.WorkflowExecutionInfo {
	var source *visibilityRecord
	err := json.Unmarshal(*hit.Source, &source)
	if err != nil { // log and skip error
		v.logger.WithFields(bark.Fields{
			"error": err.Error(),
			"docID": hit.Id,
		}).Error("unable to unmarshal search hit source")
		return nil
	}

	execution := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(source.WorkflowID),
		RunId:      common.StringPtr(source.RunID),
	}
	wfType := &workflow.WorkflowType{
		Name: common.StringPtr(source.WorkflowType),
	}

	var record *workflow.WorkflowExecutionInfo
	if isOpen {
		record = &workflow.WorkflowExecutionInfo{
			Execution: execution,
			Type:      wfType,
			StartTime: common.Int64Ptr(source.StartTime),
		}
	} else {
		record = &workflow.WorkflowExecutionInfo{
			Execution:     execution,
			Type:          wfType,
			StartTime:     common.Int64Ptr(source.StartTime),
			CloseTime:     common.Int64Ptr(source.CloseTime),
			CloseStatus:   &source.CloseStatus,
			HistoryLength: common.Int64Ptr(source.HistoryLength),
		}
	}
	return record
}
