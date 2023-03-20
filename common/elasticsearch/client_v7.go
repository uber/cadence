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
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/olivere/elastic/v7"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/elasticsearch/query"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

var _ GenericClient = (*elasticV7)(nil)

type (
	// elasticV7 implements Client
	elasticV7 struct {
		client *elastic.Client
		logger log.Logger
	}

	// searchParametersV7 holds all required and optional parameters for executing a search
	searchParametersV7 struct {
		Index       string
		Query       query.Query
		From        int
		PageSize    int
		Sorter      []elastic.Sorter
		SearchAfter []interface{}
	}
)

// NewV7Client returns a new implementation of GenericClient
func NewV7Client(
	connectConfig *config.ElasticSearchConfig,
	tlsClient *http.Client,
	awsSigningClient *http.Client,
	logger log.Logger,
	clientOptFuncs ...elastic.ClientOptionFunc,
) (GenericClient, error) {
	clientOptFuncs = append(clientOptFuncs,
		elastic.SetURL(connectConfig.URL.String()),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(128*time.Millisecond, 513*time.Millisecond))),
		elastic.SetDecoder(&elastic.NumberDecoder{}), // critical to ensure decode of int64 won't lose precise
	)
	if connectConfig.DisableSniff {
		clientOptFuncs = append(clientOptFuncs, elastic.SetSniff(false))
	}
	if connectConfig.DisableHealthCheck {
		clientOptFuncs = append(clientOptFuncs, elastic.SetHealthcheck(false))
	}

	if awsSigningClient != nil {
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(awsSigningClient))
	}

	if tlsClient != nil {
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(tlsClient))
	}

	client, err := elastic.NewClient(clientOptFuncs...)
	if err != nil {
		return nil, err
	}

	return &elasticV7{
		client: client,
		logger: logger,
	}, nil
}

func (c *elasticV7) IsNotFoundError(err error) bool {
	return elastic.IsNotFound(err)
}

// root is for nested object like Attr property for search attributes.
func (c *elasticV7) PutMapping(ctx context.Context, index, root, key, valueType string) error {
	body := buildPutMappingBodyV7(root, key, valueType)
	_, err := c.client.PutMapping().Index(index).BodyJson(body).Do(ctx)
	return err
}

func (c *elasticV7) CreateIndex(ctx context.Context, index string) error {
	_, err := c.client.CreateIndex(index).Do(ctx)
	return err
}

func (c *elasticV7) CountByQuery(ctx context.Context, index, query string) (int64, error) {
	return c.client.Count(index).BodyString(query).Do(ctx)
}

func (c *elasticV7) Search(ctx context.Context, request *SearchRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	token, err := GetNextPageToken(request.ListRequest.NextPageToken)
	if err != nil {
		return nil, err
	}

	var matchQuery *query.MatchQuery
	if request.MatchQuery != nil {
		matchQuery = query.NewMatchQuery(request.MatchQuery.Name, request.MatchQuery.Text)
	}

	searchResult, err := c.getSearchResult(
		ctx,
		request.Index,
		request.ListRequest,
		matchQuery,
		request.IsOpen,
		token,
	)

	if err != nil {
		return nil, err
	}

	return c.getListWorkflowExecutionsResponse(searchResult.Hits, token, request.ListRequest.PageSize, request.MaxResultWindow, request.Filter)
}

func (c *elasticV7) SearchByQuery(ctx context.Context, request *SearchByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	token, err := GetNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}
	searchResult, err := c.client.Search(request.Index).Source(request.Query).Do(ctx)
	if err != nil {
		return nil, err
	}

	return c.getListWorkflowExecutionsResponse(searchResult.Hits, token, request.PageSize, request.MaxResultWindow, request.Filter)
}

func (c *elasticV7) ScanByQuery(ctx context.Context, request *ScanByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	var err error
	token, err := GetNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}
	searchResult, err := c.scroll(ctx, request.Index, request.Query, token.ScrollID)

	isLastPage := false
	if err == io.EOF { // no more result
		isLastPage = true
		if err := c.clearScroll(ctx, searchResult.ScrollId); err != nil {
			c.logger.Warn("scroll clear failed", tag.Error(err))
		}
	} else if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ScanByQuery failed. Error: %v", err),
		}
	}
	response := &p.InternalListWorkflowExecutionsResponse{}
	actualHits := searchResult.Hits.Hits
	numOfActualHits := len(actualHits)
	response.Executions = c.esHitsToExecutions(searchResult.Hits, nil /* no filter */)

	if numOfActualHits == request.PageSize && !isLastPage {
		nextPageToken, err := SerializePageToken(&ElasticVisibilityPageToken{ScrollID: searchResult.ScrollId})
		if err != nil {
			return nil, err
		}
		response.NextPageToken = make([]byte, len(nextPageToken))
		copy(response.NextPageToken, nextPageToken)
	}

	return response, nil
}

func (c *elasticV7) scroll(ctx context.Context, index, query, scrollID string) (*elastic.SearchResult, error) {
	scrollService := elastic.NewScrollService(c.client)
	if len(scrollID) == 0 {
		return scrollService.Index(index).Body(query).Do(ctx)
	}
	return scrollService.ScrollId(scrollID).Do(ctx)
}

func (c *elasticV7) clearScroll(ctx context.Context, scrollID string) error {
	return elastic.NewScrollService(c.client).ScrollId(scrollID).Clear(ctx)
}

func (c *elasticV7) SearchForOneClosedExecution(
	ctx context.Context,
	index string,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {

	matchDomainQuery := query.NewMatchQuery(DomainID, request.DomainUUID)
	existClosedStatusQuery := query.NewExistsQuery(CloseStatus)
	matchWorkflowIDQuery := query.NewMatchQuery(WorkflowID, request.Execution.GetWorkflowID())
	boolQuery := query.NewBoolQuery().Must(matchDomainQuery).Must(existClosedStatusQuery).Must(matchWorkflowIDQuery)
	rid := request.Execution.GetRunID()
	if rid != "" {
		matchRunIDQuery := query.NewMatchQuery(RunID, rid)
		boolQuery = boolQuery.Must(matchRunIDQuery)
	}

	params := &searchParametersV7{
		Index: index,
		Query: boolQuery,
	}
	searchResult, err := c.search(ctx, params)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("SearchForOneClosedExecution failed. Error: %v", err),
		}
	}

	response := &p.InternalGetClosedWorkflowExecutionResponse{}
	actualHits := searchResult.Hits.Hits
	if len(actualHits) == 0 {
		return response, nil
	}
	response.Execution = c.convertSearchResultToVisibilityRecord(actualHits[0])

	return response, nil
}

func (c *elasticV7) search(ctx context.Context, p *searchParametersV7) (*elastic.SearchResult, error) {
	searchService := c.client.Search(p.Index).
		Query(p.Query).
		From(p.From).
		SortBy(p.Sorter...)

	if p.PageSize != 0 {
		searchService.Size(p.PageSize)
	}

	if len(p.SearchAfter) != 0 {
		searchService.SearchAfter(p.SearchAfter...)
	}

	return searchService.Do(ctx)
}

func (c *elasticV7) SearchRaw(ctx context.Context, index string, query string) (*RawResponse, error) {
	// There's slight differences between the v6 and v7 response preventing us to move
	// this to a common function
	esResult, err := c.client.Search(index).Source(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	if esResult.Error != nil {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch Error: %#v", esResult.Error),
		}
	} else if esResult.TimedOut {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch Error: Request timed out: %v ms", esResult.TookInMillis),
		}
	}

	result := RawResponse{
		TookInMillis: esResult.TookInMillis,
		Hits: SearchHits{
			TotalHits: esResult.TotalHits(),
			Hits:      c.esHitsToExecutions(esResult.Hits, nil /*no filter*/),
		},
		Aggregations: esResult.Aggregations,
	}

	return &result, nil
}

func (c *elasticV7) esHitsToExecutions(eshits *elastic.SearchHits, filter IsRecordValidFilter) []*p.InternalVisibilityWorkflowExecutionInfo {
	var hits = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	if eshits != nil && len(eshits.Hits) > 0 {
		for _, hit := range eshits.Hits {
			workflowExecutionInfo := c.convertSearchResultToVisibilityRecord(hit)
			if filter == nil || filter(workflowExecutionInfo) {
				hits = append(hits, workflowExecutionInfo)
			}
		}
	}
	return hits
}

func buildPutMappingBodyV7(root, key, valueType string) map[string]interface{} {
	body := make(map[string]interface{})
	if len(root) != 0 {
		body["properties"] = map[string]interface{}{
			root: map[string]interface{}{
				"properties": map[string]interface{}{
					key: map[string]interface{}{
						"type": valueType,
					},
				},
			},
		}
	} else {
		body["properties"] = map[string]interface{}{
			key: map[string]interface{}{
				"type": valueType,
			},
		}
	}
	return body
}

func (c *elasticV7) getListWorkflowExecutionsResponse(
	searchHits *elastic.SearchHits,
	token *ElasticVisibilityPageToken,
	pageSize int,
	maxResultWindow int,
	isRecordValid func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool,
) (*p.InternalListWorkflowExecutionsResponse, error) {

	response := &p.InternalListWorkflowExecutionsResponse{}
	actualHits := searchHits.Hits
	numOfActualHits := len(actualHits)

	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo := c.convertSearchResultToVisibilityRecord(actualHits[i])
		if isRecordValid == nil || isRecordValid(workflowExecutionInfo) {
			// for old APIs like ListOpenWorkflowExecutions, we added 1 ms to range query to overcome ES limitation
			// (see getSearchResult function), but manually dropped records beyond request range here.
			response.Executions = append(response.Executions, workflowExecutionInfo)
		}
	}

	if numOfActualHits == pageSize { // this means the response is not the last page
		var nextPageToken []byte
		var err error

		// ES Search API support pagination using From and PageSize, but has limit that From+PageSize cannot exceed a threshold
		// to retrieve deeper pages, use ES SearchAfter
		if searchHits.TotalHits.Value <= int64(maxResultWindow-pageSize) { // use ES Search From+Size
			nextPageToken, err = SerializePageToken(&ElasticVisibilityPageToken{From: token.From + numOfActualHits})
		} else { // use ES Search After
			var sortVal interface{}
			sortVals := actualHits[len(response.Executions)-1].Sort
			sortVal = sortVals[0]
			tieBreaker := sortVals[1].(string)

			nextPageToken, err = SerializePageToken(&ElasticVisibilityPageToken{SortValue: sortVal, TieBreaker: tieBreaker})
		}
		if err != nil {
			return nil, err
		}

		response.NextPageToken = make([]byte, len(nextPageToken))
		copy(response.NextPageToken, nextPageToken)
	}

	return response, nil
}

func (c *elasticV7) convertSearchResultToVisibilityRecord(hit *elastic.SearchHit) *p.InternalVisibilityWorkflowExecutionInfo {
	var source *VisibilityRecord
	err := json.Unmarshal(hit.Source, &source)
	if err != nil { // log and skip error
		c.logger.Error("unable to unmarshal search hit source",
			tag.Error(err), tag.ESDocID(hit.Id))
		return nil
	}

	record := &p.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         source.DomainID,
		WorkflowType:     source.WorkflowType,
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.Unix(0, source.StartTime),
		ExecutionTime:    time.Unix(0, source.ExecutionTime),
		Memo:             p.NewDataBlob(source.Memo, common.EncodingType(source.Encoding)),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		NumClusters:      source.NumClusters,
		SearchAttributes: source.Attr,
	}
	if source.UpdateTime != 0 {
		record.UpdateTime = time.Unix(0, source.UpdateTime)
	}
	if source.CloseTime != 0 {
		record.CloseTime = time.Unix(0, source.CloseTime)
		record.Status = thrift.ToWorkflowExecutionCloseStatus(&source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func (c *elasticV7) getSearchResult(
	ctx context.Context,
	index string,
	request *p.InternalListWorkflowExecutionsRequest,
	matchQuery *query.MatchQuery,
	isOpen bool,
	token *ElasticVisibilityPageToken,
) (*elastic.SearchResult, error) {

	// always match domain id
	//boolQuery := elastic.NewBoolQuery().Must(elastic.NewMatchQuery(DomainID, request.DomainUUID))
	boolQuery := query.NewBoolQuery().Must(query.NewMatchQuery(DomainID, request.DomainUUID))
	if matchQuery != nil {
		boolQuery = boolQuery.Must(matchQuery)
	}
	// ElasticSearch v6 is unable to precisely compare time, have to manually add resolution 1ms to time range.
	// Also has to use string instead of int64 to avoid data conversion issue,
	// 9223372036854775807 to 9223372036854776000 (long overflow)
	if request.LatestTime.UnixNano() > math.MaxInt64-oneMicroSecondInNano { // prevent latestTime overflow
		request.LatestTime = time.Unix(0, math.MaxInt64-oneMicroSecondInNano)
	}
	if request.EarliestTime.UnixNano() < math.MinInt64+oneMicroSecondInNano { // prevent earliestTime overflow
		request.EarliestTime = time.Unix(0, math.MinInt64+oneMicroSecondInNano)
	}

	var rangeQueryField string
	existClosedStatusQuery := elastic.NewExistsQuery(CloseStatus)
	if isOpen {
		rangeQueryField = StartTime
		boolQuery = boolQuery.MustNot(existClosedStatusQuery)
	} else {
		boolQuery = boolQuery.Must(existClosedStatusQuery)
		rangeQueryField = CloseTime
	}

	earliestTimeStr := strconv.FormatInt(request.EarliestTime.UnixNano()-oneMicroSecondInNano, 10)
	latestTimeStr := strconv.FormatInt(request.LatestTime.UnixNano()+oneMicroSecondInNano, 10)
	rangeQuery := elastic.NewRangeQuery(rangeQueryField).Gte(earliestTimeStr).Lte(latestTimeStr)
	boolQuery = boolQuery.Filter(rangeQuery)

	params := &searchParametersV7{
		Index:    index,
		Query:    boolQuery,
		From:     token.From,
		PageSize: request.PageSize,
	}

	params.Sorter = append(params.Sorter, query.NewFieldSort(rangeQueryField).Desc())
	params.Sorter = append(params.Sorter, query.NewFieldSort(RunID).Desc())

	if ShouldSearchAfter(token) {
		params.SearchAfter = []interface{}{token.SortValue, token.TieBreaker}
	}
	return c.search(ctx, params)
}
