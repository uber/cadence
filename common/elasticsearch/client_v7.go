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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/olivere/elastic/v7"
	esaws "github.com/olivere/elastic/v7/aws/v4"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

var _ GenericClient = (*elasticV7)(nil)
var _ GenericBulkProcessor = (*v7BulkProcessor)(nil)

type (
	// elasticV7 implements Client
	elasticV7 struct {
		client     *elastic.Client
		logger     log.Logger
		serializer p.PayloadSerializer
	}

	// searchParametersV7 holds all required and optional parameters for executing a search
	searchParametersV7 struct {
		Index       string
		Query       elastic.Query
		From        int
		PageSize    int
		Sorter      []elastic.Sorter
		SearchAfter []interface{}
	}

	// bulkProcessorParametersV7 holds all required and optional parameters for executing bulk service
	bulkProcessorParametersV7 struct {
		Name          string
		NumOfWorkers  int
		BulkActions   int
		BulkSize      int
		FlushInterval time.Duration
		Backoff       elastic.Backoff
		BeforeFunc    elastic.BulkBeforeFunc
		AfterFunc     elastic.BulkAfterFunc
	}
)

// NewV7Client returns a new implementation of GenericClient
func NewV7Client(
	connectConfig *config.ElasticSearchConfig,
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
	if connectConfig.AWSSigning.Enable {
		if err := config.CheckAWSSigningConfig(connectConfig.AWSSigning); err != nil {
			return nil, err
		}
		var signingClient *http.Client
		var err error
		if connectConfig.AWSSigning.EnvironmentCredential != nil {
			signingClient, err = buildSigningHTTPClientFromEnvironmentCredentialV7(*connectConfig.AWSSigning.EnvironmentCredential)
		} else {
			signingClient, err = buildSigningHTTPClientFromStaticCredentialV7(*connectConfig.AWSSigning.StaticCredential)
		}
		if err != nil {
			return nil, err
		}
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(signingClient))
	}
	if connectConfig.TLS.Enabled {
		var tlsClient *http.Client
		var err error
		tlsClient, err = buildTLSHTTPClient(connectConfig.TLS)
		if err != nil {
			return nil, err
		}
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(tlsClient))
	}
	client, err := elastic.NewClient(clientOptFuncs...)
	if err != nil {
		return nil, err
	}

	return &elasticV7{
		client:     client,
		logger:     logger,
		serializer: p.NewPayloadSerializer(),
	}, nil
}

// refer to https://github.com/olivere/elastic/blob/release-branch.v7/recipes/aws-connect-v4/main.go
func buildSigningHTTPClientFromStaticCredentialV7(credentialConfig config.AWSStaticCredential) (*http.Client, error) {
	awsCredentials := credentials.NewStaticCredentials(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
		credentialConfig.SessionToken,
	)
	return esaws.NewV4SigningClient(awsCredentials, credentialConfig.Region), nil
}

func buildSigningHTTPClientFromEnvironmentCredentialV7(credentialConfig config.AWSEnvironmentCredential) (*http.Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(credentialConfig.Region)},
	)
	if err != nil {
		return nil, err
	}
	return esaws.NewV4SigningClient(sess.Config.Credentials, credentialConfig.Region), nil
}

func (c *elasticV7) IsNotFoundError(err error) bool {
	if elastic.IsNotFound(err) {
		return true
	}
	return false
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

	var matchQuery *elastic.MatchQuery
	if request.MatchQuery != nil {
		matchQuery = elastic.NewMatchQuery(request.MatchQuery.Name, request.MatchQuery.Text)
	}

	token, err := GetNextPageToken(request.ListRequest.NextPageToken)
	if err != nil {
		return nil, err
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
	searchResult, err := c.searchWithDSL(ctx, request.Index, request.Query)
	if err != nil {
		return nil, err
	}

	token, err := GetNextPageToken(request.NextPageToken)
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

	var searchResult *elastic.SearchResult
	var scrollService *elastic.ScrollService

	if len(token.ScrollID) == 0 { // first call
		searchResult, scrollService, err = c.scrollFirstPage(ctx, request.Index, request.Query)
	} else {
		searchResult, scrollService, err = c.scroll(ctx, token.ScrollID)
	}

	isLastPage := false
	if err == io.EOF { // no more result
		isLastPage = true
		if scrollService != nil {
			err := scrollService.Clear(ctx)
			if err != nil {
				c.logger.Warn("scrollService Clear fail", tag.Error(err))
			}
		}
	} else if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ScanByQuery failed. Error: %v", err),
		}
	}

	return c.getScanWorkflowExecutionsResponse(searchResult.Hits, request.PageSize, searchResult.ScrollId, isLastPage)
}

func (c *elasticV7) RunBulkProcessor(ctx context.Context, parameters *BulkProcessorParameters) (GenericBulkProcessor, error) {
	beforeFunc := func(executionId int64, requests []elastic.BulkableRequest) {
		parameters.BeforeFunc(executionId, fromV7ToGenericBulkableRequests(requests))
	}

	afterFunc := func(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
		gerr := convertV7ErrorToGenericError(err)
		parameters.AfterFunc(
			executionId,
			fromV7ToGenericBulkableRequests(requests),
			fromV7toGenericBulkResponse(response),
			gerr)
	}

	return c.runBulkProcessor(ctx, &bulkProcessorParametersV7{
		Name:          parameters.Name,
		NumOfWorkers:  parameters.NumOfWorkers,
		BulkActions:   parameters.BulkActions,
		BulkSize:      parameters.BulkSize,
		FlushInterval: parameters.FlushInterval,
		Backoff:       parameters.Backoff,
		BeforeFunc:    beforeFunc,
		AfterFunc:     afterFunc,
	})
}

func convertV7ErrorToGenericError(err error) *GenericError {
	if err == nil {
		return nil
	}
	status := unknownStatusCode
	switch e := err.(type) {
	case *elastic.Error:
		status = e.Status
	}
	return &GenericError{
		Status:  status,
		Details: err,
	}
}

func (v *v7BulkProcessor) RetrieveKafkaKey(request GenericBulkableRequest, logger log.Logger, metricsClient metrics.Client) string {
	req, err := request.Source()
	if err != nil {
		logger.Error("Get request source err.", tag.Error(err), tag.ESRequest(request.String()))
		metricsClient.IncCounter(metrics.ESProcessorScope, metrics.ESProcessorCorruptedData)
		return ""
	}

	var key string
	if len(req) == 2 { // index or update requests
		var body map[string]interface{}
		if err := json.Unmarshal([]byte(req[1]), &body); err != nil {
			logger.Error("Unmarshal index request body err.", tag.Error(err))
			metricsClient.IncCounter(metrics.ESProcessorScope, metrics.ESProcessorCorruptedData)
			return ""
		}

		k, ok := body[KafkaKey]
		if !ok {
			// must be bug in code and bad deployment, check processor that add es requests
			panic("KafkaKey not found")
		}
		key, ok = k.(string)
		if !ok {
			// must be bug in code and bad deployment, check processor that add es requests
			panic("KafkaKey is not string")
		}
	} else { // delete requests
		var body map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(req[0]), &body); err != nil {
			logger.Error("Unmarshal delete request body err.", tag.Error(err))
			metricsClient.IncCounter(metrics.ESProcessorScope, metrics.ESProcessorCorruptedData)
			return ""
		}

		opMap, ok := body["delete"]
		if !ok {
			// must be bug, check if dependency changed
			panic("delete key not found in request")
		}
		k, ok := opMap["_id"]
		if !ok {
			// must be bug in code and bad deployment, check processor that add es requests
			panic("_id not found in request opMap")
		}
		key, _ = k.(string)
	}
	return key
}

func (c *elasticV7) SearchForOneClosedExecution(
	ctx context.Context,
	index string,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {

	matchDomainQuery := elastic.NewMatchQuery(DomainID, request.DomainUUID)
	existClosedStatusQuery := elastic.NewExistsQuery(CloseStatus)
	matchWorkflowIDQuery := elastic.NewMatchQuery(WorkflowID, request.Execution.GetWorkflowID())
	boolQuery := elastic.NewBoolQuery().Must(matchDomainQuery).Must(existClosedStatusQuery).Must(matchWorkflowIDQuery)
	rid := request.Execution.GetRunID()
	if rid != "" {
		matchRunIDQuery := elastic.NewMatchQuery(RunID, rid)
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

func fromV7toGenericBulkResponse(response *elastic.BulkResponse) *GenericBulkResponse {
	return &GenericBulkResponse{
		Took:   response.Took,
		Errors: response.Errors,
		Items:  fromV7ToGenericBulkResponseItemMaps(response.Items),
	}
}

func fromV7ToGenericBulkResponseItemMaps(items []map[string]*elastic.BulkResponseItem) []map[string]*GenericBulkResponseItem {
	var gitems []map[string]*GenericBulkResponseItem
	for _, it := range items {
		gitems = append(gitems, fromV7ToGenericBulkResponseItemMap(it))
	}
	return gitems
}

func fromV7ToGenericBulkResponseItemMap(m map[string]*elastic.BulkResponseItem) map[string]*GenericBulkResponseItem {
	if m == nil {
		return nil
	}
	gm := make(map[string]*GenericBulkResponseItem, len(m))
	for k, v := range m {
		gm[k] = fromV7ToGenericBulkResponseItem(v)
	}
	return gm
}

func fromV7ToGenericBulkResponseItem(v *elastic.BulkResponseItem) *GenericBulkResponseItem {
	return &GenericBulkResponseItem{
		Index:         v.Index,
		Type:          v.Type,
		ID:            v.Id,
		Version:       v.Version,
		Result:        v.Result,
		SeqNo:         v.SeqNo,
		PrimaryTerm:   v.PrimaryTerm,
		Status:        v.Status,
		ForcedRefresh: v.ForcedRefresh,
	}
}

func fromV7ToGenericBulkableRequests(requests []elastic.BulkableRequest) []GenericBulkableRequest {
	var v7Reqs []GenericBulkableRequest
	for _, req := range requests {
		v7Reqs = append(v7Reqs, req)
	}
	return v7Reqs
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

func (c *elasticV7) searchWithDSL(ctx context.Context, index, query string) (*elastic.SearchResult, error) {
	return c.client.Search(index).Source(query).Do(ctx)
}

func (c *elasticV7) scroll(ctx context.Context, scrollID string) (
	*elastic.SearchResult, *elastic.ScrollService, error) {

	scrollService := elastic.NewScrollService(c.client)
	result, err := scrollService.ScrollId(scrollID).Do(ctx)
	return result, scrollService, err
}

func (c *elasticV7) scrollFirstPage(ctx context.Context, index, query string) (
	*elastic.SearchResult, *elastic.ScrollService, error) {

	scrollService := elastic.NewScrollService(c.client)
	result, err := scrollService.Index(index).Body(query).Do(ctx)
	return result, scrollService, err
}

type v7BulkProcessor struct {
	processor *elastic.BulkProcessor
}

func (v *v7BulkProcessor) Start(ctx context.Context) error {
	return v.processor.Start(ctx)
}

func (v *v7BulkProcessor) Stop() error {
	return v.processor.Stop()
}

func (v *v7BulkProcessor) Close() error {
	return v.processor.Close()
}

func (v *v7BulkProcessor) Add(request *GenericBulkableAddRequest) {
	var req elastic.BulkableRequest
	if request.IsDelete {
		req = elastic.NewBulkDeleteRequest().
			Index(request.Index).
			Type(request.Type).
			Id(request.ID).
			VersionType(request.VersionType).
			Version(request.Version)
	} else {
		req = elastic.NewBulkIndexRequest().
			Index(request.Index).
			Type(request.Type).
			Id(request.ID).
			VersionType(request.VersionType).
			Version(request.Version).
			Doc(request.Doc)
	}
	v.processor.Add(req)
}

func (v *v7BulkProcessor) Flush() error {
	return v.processor.Flush()
}

func (c *elasticV7) runBulkProcessor(ctx context.Context, p *bulkProcessorParametersV7) (*v7BulkProcessor, error) {
	processor, err := c.client.BulkProcessor().
		Name(p.Name).
		Workers(p.NumOfWorkers).
		BulkActions(p.BulkActions).
		BulkSize(p.BulkSize).
		FlushInterval(p.FlushInterval).
		Backoff(p.Backoff).
		Before(p.BeforeFunc).
		After(p.AfterFunc).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return &v7BulkProcessor{
		processor: processor,
	}, nil
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
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.Unix(0, source.StartTime),
		ExecutionTime:    time.Unix(0, source.ExecutionTime),
		Memo:             p.NewDataBlob(source.Memo, common.EncodingType(source.Encoding)),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		SearchAttributes: source.Attr,
	}
	if source.CloseTime != 0 {
		record.CloseTime = time.Unix(0, source.CloseTime)
		record.Status = thrift.ToWorkflowExecutionCloseStatus(&source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func (c *elasticV7) getScanWorkflowExecutionsResponse(
	searchHits *elastic.SearchHits,
	pageSize int, scrollID string,
	isLastPage bool,
) (*p.InternalListWorkflowExecutionsResponse, error) {

	response := &p.InternalListWorkflowExecutionsResponse{}
	actualHits := searchHits.Hits
	numOfActualHits := len(actualHits)

	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo := c.convertSearchResultToVisibilityRecord(actualHits[i])
		response.Executions = append(response.Executions, workflowExecutionInfo)
	}

	if numOfActualHits == pageSize && !isLastPage {
		nextPageToken, err := SerializePageToken(&ElasticVisibilityPageToken{ScrollID: scrollID})
		if err != nil {
			return nil, err
		}
		response.NextPageToken = make([]byte, len(nextPageToken))
		copy(response.NextPageToken, nextPageToken)
	}

	return response, nil
}

func (c *elasticV7) getSearchResult(
	ctx context.Context,
	index string,
	request *p.InternalListWorkflowExecutionsRequest,
	matchQuery *elastic.MatchQuery,
	isOpen bool,
	token *ElasticVisibilityPageToken,
) (*elastic.SearchResult, error) {

	matchDomainQuery := elastic.NewMatchQuery(DomainID, request.DomainUUID)
	existClosedStatusQuery := elastic.NewExistsQuery(CloseStatus)
	var rangeQuery *elastic.RangeQuery
	if isOpen {
		rangeQuery = elastic.NewRangeQuery(StartTime)
	} else {
		rangeQuery = elastic.NewRangeQuery(CloseTime)
	}
	// ElasticSearch v7 is unable to precisely compare time, have to manually add resolution 1ms to time range.
	// Also has to use string instead of int64 to avoid data conversion issue,
	// 9223372036854775807 to 9223372036854776000 (long overflow)
	if request.LatestTime.UnixNano() > math.MaxInt64-oneMicroSecondInNano { // prevent latestTime overflow
		request.LatestTime = time.Unix(0, math.MaxInt64-oneMicroSecondInNano)
	}
	if request.EarliestTime.UnixNano() < math.MinInt64+oneMicroSecondInNano { // prevent earliestTime overflow
		request.EarliestTime = time.Unix(0, math.MinInt64+oneMicroSecondInNano)
	}
	earliestTimeStr := strconv.FormatInt(request.EarliestTime.UnixNano()-oneMicroSecondInNano, 10)
	latestTimeStr := strconv.FormatInt(request.LatestTime.UnixNano()+oneMicroSecondInNano, 10)
	rangeQuery = rangeQuery.
		Gte(earliestTimeStr).
		Lte(latestTimeStr)

	boolQuery := elastic.NewBoolQuery().Must(matchDomainQuery).Filter(rangeQuery)
	if matchQuery != nil {
		boolQuery = boolQuery.Must(matchQuery)
	}
	if isOpen {
		boolQuery = boolQuery.MustNot(existClosedStatusQuery)
	} else {
		boolQuery = boolQuery.Must(existClosedStatusQuery)
	}

	params := &searchParametersV7{
		Index:    index,
		Query:    boolQuery,
		From:     token.From,
		PageSize: request.PageSize,
	}
	if isOpen {
		params.Sorter = append(params.Sorter, elastic.NewFieldSort(StartTime).Desc())
	} else {
		params.Sorter = append(params.Sorter, elastic.NewFieldSort(CloseTime).Desc())
	}
	params.Sorter = append(params.Sorter, elastic.NewFieldSort(RunID).Desc())

	if ShouldSearchAfter(token) {
		params.SearchAfter = []interface{}{token.SortValue, token.TieBreaker}
	}

	return c.search(ctx, params)
}
