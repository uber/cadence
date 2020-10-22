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

package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"

	"github.com/cch123/elasticsql"
	"github.com/valyala/fastjson"

	"github.com/uber/cadence/.gen/go/indexer"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"
)

const (
	esPersistenceName = "elasticsearch"
)

type (
	esVisibilityStore struct {
		esClient   es.GenericClient
		index      string
		producer   messaging.Producer
		logger     log.Logger
		config     *config.VisibilityConfig
		serializer p.PayloadSerializer
	}
)

var _ p.VisibilityStore = (*esVisibilityStore)(nil)

// NewElasticSearchVisibilityStore create a visibility store connecting to ElasticSearch
func NewElasticSearchVisibilityStore(
	esClient es.GenericClient,
	index string,
	producer messaging.Producer,
	config *config.VisibilityConfig,
	logger log.Logger,
) p.VisibilityStore {
	return &esVisibilityStore{
		esClient:   esClient,
		index:      index,
		producer:   producer,
		logger:     logger.WithTags(tag.ComponentESVisibilityManager),
		config:     config,
		serializer: p.NewPayloadSerializer(),
	}
}

func (v *esVisibilityStore) Close() {}

func (v *esVisibilityStore) GetName() string {
	return esPersistenceName
}

func (v *esVisibilityStore) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	v.checkProducer()
	memo := v.serializeMemo(request.Memo, request.DomainUUID, request.WorkflowID, request.RunID)
	msg := getVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.TaskList,
		request.StartTimestamp,
		request.ExecutionTimestamp,
		request.TaskID,
		memo.Data,
		memo.GetEncoding(),
		request.SearchAttributes,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *esVisibilityStore) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionClosedRequest,
) error {
	v.checkProducer()
	memo := v.serializeMemo(request.Memo, request.DomainUUID, request.WorkflowID, request.RunID)
	msg := getVisibilityMessageForCloseExecution(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.StartTimestamp,
		request.ExecutionTimestamp,
		request.CloseTimestamp,
		*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
		request.HistoryLength,
		request.TaskID,
		memo.Data,
		request.TaskList,
		memo.GetEncoding(),
		request.SearchAttributes,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *esVisibilityStore) UpsertWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpsertWorkflowExecutionRequest,
) error {
	v.checkProducer()
	memo := v.serializeMemo(request.Memo, request.DomainUUID, request.WorkflowID, request.RunID)
	msg := getVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.TaskList,
		request.StartTimestamp,
		request.ExecutionTimestamp,
		request.TaskID,
		memo.Data,
		memo.GetEncoding(),
		request.SearchAttributes,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *esVisibilityStore) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		startTime := rec.StartTime.UnixNano()
		return request.EarliestTime <= startTime && startTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: request,
		IsOpen:      true,
		Filter:      isRecordValid,
		MatchQuery:  nil,
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutions failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		closeTime := rec.CloseTime.UnixNano()
		return request.EarliestTime <= closeTime && closeTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: request,
		IsOpen:      false,
		Filter:      isRecordValid,
		MatchQuery:  nil,
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		startTime := rec.StartTime.UnixNano()
		return request.EarliestTime <= startTime && startTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: &request.InternalListWorkflowExecutionsRequest,
		IsOpen:      true,
		Filter:      isRecordValid,
		MatchQuery: &es.GenericMatch{
			Name: es.WorkflowType,
			Text: request.WorkflowTypeName,
		},
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutionsByType failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		closeTime := rec.CloseTime.UnixNano()
		return request.EarliestTime <= closeTime && closeTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: &request.InternalListWorkflowExecutionsRequest,
		IsOpen:      false,
		Filter:      isRecordValid,
		MatchQuery: &es.GenericMatch{
			Name: es.WorkflowType,
			Text: request.WorkflowTypeName,
		},
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByType failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		startTime := rec.StartTime.UnixNano()
		return request.EarliestTime <= startTime && startTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: &request.InternalListWorkflowExecutionsRequest,
		IsOpen:      true,
		Filter:      isRecordValid,
		MatchQuery: &es.GenericMatch{
			Name: es.WorkflowID,
			Text: request.WorkflowID,
		},
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListOpenWorkflowExecutionsByWorkflowID failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		closeTime := rec.CloseTime.UnixNano()
		return request.EarliestTime <= closeTime && closeTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: &request.InternalListWorkflowExecutionsRequest,
		IsOpen:      false,
		Filter:      isRecordValid,
		MatchQuery: &es.GenericMatch{
			Name: es.WorkflowID,
			Text: request.WorkflowID,
		},
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByWorkflowID failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		closeTime := rec.CloseTime.UnixNano()
		return request.EarliestTime <= closeTime && closeTime <= request.LatestTime
	}

	resp, err := v.esClient.Search(ctx, &es.SearchRequest{
		Index:       v.index,
		ListRequest: &request.InternalListWorkflowExecutionsRequest,
		IsOpen:      false,
		Filter:      isRecordValid,
		MatchQuery: &es.GenericMatch{
			Name: es.CloseStatus,
			Text: int32(*thrift.FromWorkflowExecutionCloseStatus(&request.Status)),
		},
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutionsByStatus failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) GetClosedWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	resp, err := v.esClient.SearchForOneClosedExecution(ctx, v.index, request)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("SearchForOneClosedExecution failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	v.checkProducer()
	msg := getVisibilityMessageForDeletion(
		request.DomainID,
		request.WorkflowID,
		request.RunID,
		request.TaskID,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *esVisibilityStore) ListWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {

	checkPageSize(request)

	token, err := es.GetNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	queryDSL, err := v.getESQueryDSL(request, token)
	if err != nil {
		return nil, &workflow.BadRequestError{Message: fmt.Sprintf("Error when parse query: %v", err)}
	}

	resp, err := v.esClient.SearchByQuery(ctx, &es.SearchByQueryRequest{
		Index:         v.index,
		Query:         queryDSL,
		NextPageToken: request.NextPageToken,
		PageSize:      request.PageSize,
		Filter:        nil,
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListWorkflowExecutions failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) ScanWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {

	checkPageSize(request)

	token, err := es.GetNextPageToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}

	var queryDSL string
	if len(token.ScrollID) == 0 { // first call
		queryDSL, err = getESQueryDSLForScan(request)
		if err != nil {
			return nil, &workflow.BadRequestError{Message: fmt.Sprintf("Error when parse query: %v", err)}
		}
	}

	resp, err := v.esClient.ScanByQuery(ctx, &es.ScanByQueryRequest{
		Index:         v.index,
		Query:         queryDSL,
		NextPageToken: request.NextPageToken,
		PageSize:      request.PageSize,
	})
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ScanWorkflowExecutions failed, %v", err),
		}
	}
	return resp, nil
}

func (v *esVisibilityStore) CountWorkflowExecutions(
	ctx context.Context,
	request *p.CountWorkflowExecutionsRequest,
) (
	*p.CountWorkflowExecutionsResponse, error) {

	queryDSL, err := getESQueryDSLForCount(request)
	if err != nil {
		return nil, &workflow.BadRequestError{Message: fmt.Sprintf("Error when parse query: %v", err)}
	}

	count, err := v.esClient.CountByQuery(ctx, v.index, queryDSL)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CountWorkflowExecutions failed. Error: %v", err),
		}
	}

	response := &p.CountWorkflowExecutionsResponse{Count: count}
	return response, nil
}

const (
	jsonMissingCloseTime     = `{"missing":{"field":"CloseTime"}}`
	jsonRangeOnExecutionTime = `{"range":{"ExecutionTime":`
	jsonSortForOpen          = `[{"StartTime":"desc"},{"RunID":"desc"}]`
	jsonSortWithTieBreaker   = `{"RunID":"desc"}`

	dslFieldSort        = "sort"
	dslFieldSearchAfter = "search_after"
	dslFieldFrom        = "from"
	dslFieldSize        = "size"

	defaultDateTimeFormat = time.RFC3339 // used for converting UnixNano to string like 2018-02-15T16:16:36-08:00
)

var (
	timeKeys = map[string]bool{
		"StartTime":     true,
		"CloseTime":     true,
		"ExecutionTime": true,
	}
	rangeKeys = map[string]bool{
		"from":  true,
		"to":    true,
		"gt":    true,
		"lt":    true,
		"query": true,
	}
)

func getESQueryDSLForScan(request *p.ListWorkflowExecutionsByQueryRequest) (string, error) {
	sql := getSQLFromListRequest(request)
	dsl, err := getCustomizedDSLFromSQL(sql, request.DomainUUID)
	if err != nil {
		return "", err
	}

	// remove not needed fields
	dsl.Del(dslFieldSort)
	return dsl.String(), nil
}

func getESQueryDSLForCount(request *p.CountWorkflowExecutionsRequest) (string, error) {
	sql := getSQLFromCountRequest(request)
	dsl, err := getCustomizedDSLFromSQL(sql, request.DomainUUID)
	if err != nil {
		return "", err
	}

	// remove not needed fields
	dsl.Del(dslFieldFrom)
	dsl.Del(dslFieldSize)
	dsl.Del(dslFieldSort)

	return dsl.String(), nil
}

func (v *esVisibilityStore) getESQueryDSL(request *p.ListWorkflowExecutionsByQueryRequest, token *es.ElasticVisibilityPageToken) (string, error) {
	sql := getSQLFromListRequest(request)
	dsl, err := getCustomizedDSLFromSQL(sql, request.DomainUUID)
	if err != nil {
		return "", err
	}

	sortField, err := v.processSortField(dsl)
	if err != nil {
		return "", err
	}

	if es.ShouldSearchAfter(token) {
		valueOfSearchAfter, err := v.getValueOfSearchAfterInJSON(token, sortField)
		if err != nil {
			return "", err
		}
		dsl.Set(dslFieldSearchAfter, fastjson.MustParse(valueOfSearchAfter))
	} else { // use from+size
		dsl.Set(dslFieldFrom, fastjson.MustParse(strconv.Itoa(token.From)))
	}

	dslStr := cleanDSL(dsl.String())

	return dslStr, nil
}

func getSQLFromListRequest(request *p.ListWorkflowExecutionsByQueryRequest) string {
	var sql string
	query := strings.TrimSpace(request.Query)
	if query == "" {
		sql = fmt.Sprintf("select * from dummy limit %d", request.PageSize)
	} else if common.IsJustOrderByClause(query) {
		sql = fmt.Sprintf("select * from dummy %s limit %d", request.Query, request.PageSize)
	} else {
		sql = fmt.Sprintf("select * from dummy where %s limit %d", request.Query, request.PageSize)
	}
	return sql
}

func getSQLFromCountRequest(request *p.CountWorkflowExecutionsRequest) string {
	var sql string
	if strings.TrimSpace(request.Query) == "" {
		sql = "select * from dummy"
	} else {
		sql = fmt.Sprintf("select * from dummy where %s", request.Query)
	}
	return sql
}

func getCustomizedDSLFromSQL(sql string, domainID string) (*fastjson.Value, error) {
	dslStr, _, err := elasticsql.Convert(sql)
	if err != nil {
		return nil, err
	}
	dsl, err := fastjson.Parse(dslStr) // dsl.String() will be a compact json without spaces
	if err != nil {
		return nil, err
	}
	dslStr = dsl.String()
	if strings.Contains(dslStr, jsonMissingCloseTime) { // isOpen
		dsl = replaceQueryForOpen(dsl)
	}
	if strings.Contains(dslStr, jsonRangeOnExecutionTime) {
		addQueryForExecutionTime(dsl)
	}
	addDomainToQuery(dsl, domainID)
	if err := processAllValuesForKey(dsl, timeKeyFilter, timeProcessFunc); err != nil {
		return nil, err
	}
	return dsl, nil
}

// ES v6 only accepts "must_not exists" query instead of "missing" query, but elasticsql produces "missing",
// so use this func to replace.
// Note it also means a temp limitation that we cannot support field missing search
func replaceQueryForOpen(dsl *fastjson.Value) *fastjson.Value {
	re := regexp.MustCompile(jsonMissingCloseTime)
	newDslStr := re.ReplaceAllString(dsl.String(), `{"bool":{"must_not":{"exists":{"field":"CloseTime"}}}}`)
	dsl = fastjson.MustParse(newDslStr)
	return dsl
}

func addQueryForExecutionTime(dsl *fastjson.Value) {
	executionTimeQueryString := `{"range" : {"ExecutionTime" : {"gt" : "0"}}}`
	addMustQuery(dsl, executionTimeQueryString)
}

func addDomainToQuery(dsl *fastjson.Value, domainID string) {
	if len(domainID) == 0 {
		return
	}

	domainQueryString := fmt.Sprintf(`{"match_phrase":{"DomainID":{"query":"%s"}}}`, domainID)
	addMustQuery(dsl, domainQueryString)
}

// addMustQuery is wrapping bool query with new bool query with must,
// reason not making a flat bool query is to ensure "should (or)" query works correctly in query context.
func addMustQuery(dsl *fastjson.Value, queryString string) {
	valOfTopQuery := dsl.Get("query")
	valOfBool := dsl.Get("query", "bool")
	newValOfBool := fmt.Sprintf(`{"must":[%s,{"bool":%s}]}`, queryString, valOfBool.String())
	valOfTopQuery.Set("bool", fastjson.MustParse(newValOfBool))
}

func (v *esVisibilityStore) processSortField(dsl *fastjson.Value) (string, error) {
	isSorted := dsl.Exists(dslFieldSort)
	var sortField string

	if !isSorted { // set default sorting by StartTime desc
		dsl.Set(dslFieldSort, fastjson.MustParse(jsonSortForOpen))
		sortField = definition.StartTime
	} else { // user provide sorting using order by
		// sort validation on length
		if len(dsl.GetArray(dslFieldSort)) > 1 {
			return "", errors.New("only one field can be used to sort")
		}
		// sort validation to exclude IndexedValueTypeString
		obj, _ := dsl.GetArray(dslFieldSort)[0].Object()
		obj.Visit(func(k []byte, v *fastjson.Value) { // visit is only way to get object key in fastjson
			sortField = string(k)
		})
		if v.getFieldType(sortField) == workflow.IndexedValueTypeString {
			return "", errors.New("not able to sort by IndexedValueTypeString field, use IndexedValueTypeKeyword field")
		}
		// add RunID as tie-breaker
		dsl.Get(dslFieldSort).Set("1", fastjson.MustParse(jsonSortWithTieBreaker))
	}

	return sortField, nil
}

func (v *esVisibilityStore) getFieldType(fieldName string) workflow.IndexedValueType {
	if strings.HasPrefix(fieldName, definition.Attr) {
		fieldName = fieldName[len(definition.Attr)+1:] // remove prefix
	}
	validMap := v.config.ValidSearchAttributes()
	fieldType, ok := validMap[fieldName]
	if !ok {
		v.logger.Error("Unknown fieldName, validation should be done in frontend already", tag.Value(fieldName))
	}
	return common.ConvertIndexedValueTypeToThriftType(fieldType, v.logger)
}

func (v *esVisibilityStore) getValueOfSearchAfterInJSON(token *es.ElasticVisibilityPageToken, sortField string) (string, error) {
	var sortVal interface{}
	var err error
	switch v.getFieldType(sortField) {
	case workflow.IndexedValueTypeInt, workflow.IndexedValueTypeDatetime, workflow.IndexedValueTypeBool:
		sortVal, err = token.SortValue.(json.Number).Int64()
		if err != nil {
			err, ok := err.(*strconv.NumError) // field not present, ES will return big int +-9223372036854776000
			if !ok {
				return "", err
			}
			if err.Num[0] == '-' { // desc
				sortVal = math.MinInt64
			} else { // asc
				sortVal = math.MaxInt64
			}
		}
	case workflow.IndexedValueTypeDouble:
		switch token.SortValue.(type) {
		case json.Number:
			sortVal, err = token.SortValue.(json.Number).Float64()
			if err != nil {
				return "", err
			}
		case string: // field not present, ES will return "-Infinity" or "Infinity"
			sortVal = fmt.Sprintf(`"%s"`, token.SortValue.(string))
		}
	case workflow.IndexedValueTypeKeyword:
		if token.SortValue != nil {
			sortVal = fmt.Sprintf(`"%s"`, token.SortValue.(string))
		} else { // field not present, ES will return null (so token.SortValue is nil)
			sortVal = "null"
		}
	default:
		sortVal = token.SortValue
	}

	return fmt.Sprintf(`[%v, "%s"]`, sortVal, token.TieBreaker), nil
}

func (v *esVisibilityStore) checkProducer() {
	if v.producer == nil {
		// must be bug, check history setup
		panic("message producer is nil")
	}
}

func getVisibilityMessage(domainID string, wid, rid string, workflowTypeName string, taskList string,
	startTimeUnixNano, executionTimeUnixNano int64, taskID int64, memo []byte, encoding common.EncodingType,
	searchAttributes map[string][]byte) *indexer.Message {

	msgType := indexer.MessageTypeIndex
	fields := map[string]*indexer.Field{
		es.WorkflowType:  {Type: &es.FieldTypeString, StringData: common.StringPtr(workflowTypeName)},
		es.StartTime:     {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(startTimeUnixNano)},
		es.ExecutionTime: {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(executionTimeUnixNano)},
		es.TaskList:      {Type: &es.FieldTypeString, StringData: common.StringPtr(taskList)},
	}
	if len(memo) != 0 {
		fields[es.Memo] = &indexer.Field{Type: &es.FieldTypeBinary, BinaryData: memo}
		fields[es.Encoding] = &indexer.Field{Type: &es.FieldTypeString, StringData: common.StringPtr(string(encoding))}
	}
	for k, v := range searchAttributes {
		fields[k] = &indexer.Field{Type: &es.FieldTypeBinary, BinaryData: v}
	}

	msg := &indexer.Message{
		MessageType: &msgType,
		DomainID:    common.StringPtr(domainID),
		WorkflowID:  common.StringPtr(wid),
		RunID:       common.StringPtr(rid),
		Version:     common.Int64Ptr(taskID),
		Fields:      fields,
	}
	return msg
}

func getVisibilityMessageForCloseExecution(domainID string, wid, rid string, workflowTypeName string,
	startTimeUnixNano int64, executionTimeUnixNano int64, endTimeUnixNano int64, closeStatus workflow.WorkflowExecutionCloseStatus,
	historyLength int64, taskID int64, memo []byte, taskList string, encoding common.EncodingType,
	searchAttributes map[string][]byte) *indexer.Message {

	msgType := indexer.MessageTypeIndex
	fields := map[string]*indexer.Field{
		es.WorkflowType:  {Type: &es.FieldTypeString, StringData: common.StringPtr(workflowTypeName)},
		es.StartTime:     {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(startTimeUnixNano)},
		es.ExecutionTime: {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(executionTimeUnixNano)},
		es.CloseTime:     {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(endTimeUnixNano)},
		es.CloseStatus:   {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(int64(closeStatus))},
		es.HistoryLength: {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(historyLength)},
		es.TaskList:      {Type: &es.FieldTypeString, StringData: common.StringPtr(taskList)},
	}
	if len(memo) != 0 {
		fields[es.Memo] = &indexer.Field{Type: &es.FieldTypeBinary, BinaryData: memo}
		fields[es.Encoding] = &indexer.Field{Type: &es.FieldTypeString, StringData: common.StringPtr(string(encoding))}
	}
	for k, v := range searchAttributes {
		fields[k] = &indexer.Field{Type: &es.FieldTypeBinary, BinaryData: v}
	}

	msg := &indexer.Message{
		MessageType: &msgType,
		DomainID:    common.StringPtr(domainID),
		WorkflowID:  common.StringPtr(wid),
		RunID:       common.StringPtr(rid),
		Version:     common.Int64Ptr(taskID),
		Fields:      fields,
	}
	return msg
}

func getVisibilityMessageForDeletion(domainID, workflowID, runID string, docVersion int64) *indexer.Message {
	msgType := indexer.MessageTypeDelete
	msg := &indexer.Message{
		MessageType: &msgType,
		DomainID:    common.StringPtr(domainID),
		WorkflowID:  common.StringPtr(workflowID),
		RunID:       common.StringPtr(runID),
		Version:     common.Int64Ptr(docVersion),
	}
	return msg
}

func checkPageSize(request *p.ListWorkflowExecutionsByQueryRequest) {
	if request.PageSize == 0 {
		request.PageSize = 1000
	}
}

func processAllValuesForKey(dsl *fastjson.Value, keyFilter func(k string) bool,
	processFunc func(obj *fastjson.Object, key string, v *fastjson.Value) error,
) error {
	switch dsl.Type() {
	case fastjson.TypeArray:
		for _, val := range dsl.GetArray() {
			if err := processAllValuesForKey(val, keyFilter, processFunc); err != nil {
				return err
			}
		}
	case fastjson.TypeObject:
		objectVal := dsl.GetObject()
		keys := []string{}
		objectVal.Visit(func(key []byte, val *fastjson.Value) {
			keys = append(keys, string(key))
		})

		for _, key := range keys {
			var err error
			val := objectVal.Get(key)
			if keyFilter(key) {
				err = processFunc(objectVal, key, val)
			} else {
				err = processAllValuesForKey(val, keyFilter, processFunc)
			}
			if err != nil {
				return err
			}
		}
	default:
		// do nothing, since there's no key
	}
	return nil
}

func timeKeyFilter(key string) bool {
	return timeKeys[key]
}

func timeProcessFunc(obj *fastjson.Object, key string, value *fastjson.Value) error {
	return processAllValuesForKey(value, func(key string) bool {
		return rangeKeys[key]
	}, func(obj *fastjson.Object, key string, v *fastjson.Value) error {
		timeStr := string(v.GetStringBytes())

		// first check if already in int64 format
		if _, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			return nil
		}

		// try to parse time
		parsedTime, err := time.Parse(defaultDateTimeFormat, timeStr)
		if err != nil {
			return err
		}

		obj.Set(key, fastjson.MustParse(fmt.Sprintf(`"%v"`, parsedTime.UnixNano())))
		return nil
	})
}

// elasticsql may transfer `Attr.Name` to "`Attr.Name`" instead of "Attr.Name" in dsl in some operator like "between and"
// this function is used to clean up
func cleanDSL(input string) string {
	var re = regexp.MustCompile("(`)(Attr.\\w+)(`)")
	result := re.ReplaceAllString(input, `$2`)
	return result
}

func (v *esVisibilityStore) serializeMemo(visibilityMemo *types.Memo, domainID, wID, rID string) *p.DataBlob {
	memo, err := v.serializer.SerializeVisibilityMemo(thrift.FromMemo(visibilityMemo), common.EncodingTypeThriftRW)
	if err != nil {
		v.logger.WithTags(
			tag.WorkflowDomainID(domainID),
			tag.WorkflowID(wID),
			tag.WorkflowRunID(rID),
			tag.Error(err)).
			Error("Unable to encode visibility memo")
	}
	if memo == nil {
		return &p.DataBlob{}
	}
	return memo
}
