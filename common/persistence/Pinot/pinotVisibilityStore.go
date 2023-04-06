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

package pinotVisibility

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/uber/cadence/.gen/go/indexer"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	p "github.com/uber/cadence/common/persistence"
	pnt "github.com/uber/cadence/common/pinot"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

const (
	pinotPersistenceName = "pinot"
	tableName            = "cadence_visibility_pinot"

	DescendingOrder = "DESC"
	AcendingOrder   = "ASC"

	DomainID            = "DomainID"
	WorkflowID          = "WorkflowID"
	RunID               = "RunID"
	WorkflowType        = "WorkflowType"
	StartTime           = "StartTime"
	ExecutionTime       = "ExecutionTime"
	CloseTime           = "CloseTime"
	CloseStatus         = "CloseStatus"
	HistoryLength       = "HistoryLength"
	Encoding            = "Encoding"
	TaskList            = "TaskList"
	IsCron              = "IsCron"
	NumClusters         = "NumClusters"
	VisibilityOperation = "VisibilityOperation"
	UpdateTime          = "UpdateTime"
	ShardID             = "ShardID"
)

type (
	pinotVisibilityStore struct {
		pinotClient pnt.GenericClient
		producer    messaging.Producer
		logger      log.Logger
		config      *service.Config
	}

	visibilityMessage struct {
		DocID         string `json:"DocID,omitempty"`
		DomainID      string `json:"DomainID,omitempty"`
		WorkflowID    string `json:"WorkflowID,omitempty"`
		RunID         string `json:"RunID,omitempty"`
		WorkflowType  string `json:"WorkflowType,omitempty"`
		TaskList      string `json:"TaskList,omitempty"`
		StartTime     int64  `json:"StartTime,omitempty"`
		ExecutionTime int64  `json:"ExecutionTime,omitempty"`
		TaskID        int64  `json:"TaskID,omitempty"`
		IsCron        bool   `json:"IsCron,omitempty"`
		NumClusters   int16  `json:"NumClusters,omitempty"`
		Attr          string `json:"Attr,omitempty"`
		UpdateTime    int64  `json:"UpdateTime,omitempty"` // update execution,
		ShardID       int64  `json:"ShardID,omitempty"`
		// specific to certain status
		CloseTime     int64 `json:"CloseTime,omitempty"`     // close execution
		CloseStatus   int   `json:"CloseStatus"`             // close execution
		HistoryLength int64 `json:"HistoryLength,omitempty"` // close execution
	}
)

var _ p.VisibilityStore = (*pinotVisibilityStore)(nil)

func NewPinotVisibilityStore(
	pinotClient pnt.GenericClient,
	config *service.Config,
	producer messaging.Producer,
	logger log.Logger,
) p.VisibilityStore {
	return &pinotVisibilityStore{
		pinotClient: pinotClient,
		producer:    producer,
		logger:      logger.WithTags(tag.ComponentPinotVisibilityManager),
		config:      config,
	}
}

func (v *pinotVisibilityStore) Close() {
	return // TODO: need to double check what is close trace do. Does it close the client?
}

func (v *pinotVisibilityStore) GetName() string {
	return pinotPersistenceName
}

func decodeAttr(attr map[string][]byte) ([]byte, error) {
	if attr == nil {
		return nil, nil
	}

	decodedMap := make(map[string]interface{})
	for key, value := range attr {
		var val interface{}
		err := json.Unmarshal(value, &val)
		if err != nil {
			return nil, err
		}
		decodedMap[key] = val
	}
	return json.Marshal(decodedMap)
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	v.checkProducer()
	//attr, err := json.Marshal(request.SearchAttributes)
	attr, err := decodeAttr(request.SearchAttributes)
	if err != nil {
		return err
	}

	msg := createVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.TaskList,
		request.StartTimestamp.UnixMilli(),
		request.ExecutionTimestamp.UnixMilli(),
		request.TaskID,
		request.Memo.Data,
		request.Memo.GetEncoding(),
		request.IsCron,
		request.NumClusters,
		string(attr),
		common.RecordStarted,
		0,                                   // will not be used
		0,                                   // will not be used
		0,                                   // will not be used
		request.UpdateTimestamp.UnixMilli(), // will be updated when workflow execution updates
		int64(request.ShardID),
	)
	return v.producer.Publish(ctx, msg)
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionClosed(ctx context.Context, request *p.InternalRecordWorkflowExecutionClosedRequest) error {
	v.checkProducer()
	attr, err := decodeAttr(request.SearchAttributes)
	if err != nil {
		return err
	}

	msg := createVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.TaskList,
		request.StartTimestamp.UnixMilli(),
		request.ExecutionTimestamp.UnixMilli(),
		request.TaskID,
		request.Memo.Data,
		request.Memo.GetEncoding(),
		request.IsCron,
		request.NumClusters,
		string(attr),
		common.RecordClosed,
		request.CloseTimestamp.UnixMilli(),
		*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
		request.HistoryLength,
		request.UpdateTimestamp.UnixMilli(),
		int64(request.ShardID),
	)
	return v.producer.Publish(ctx, msg)
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionUninitialized(ctx context.Context, request *p.InternalRecordWorkflowExecutionUninitializedRequest) error {
	v.checkProducer()
	msg := createVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		"",
		0,
		0,
		0,
		nil,
		"",
		false,
		0,
		"",
		"",
		0,
		0,
		0,
		request.UpdateTimestamp.UnixMilli(),
		request.ShardID,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *pinotVisibilityStore) UpsertWorkflowExecution(ctx context.Context, request *p.InternalUpsertWorkflowExecutionRequest) error {
	v.checkProducer()
	attr, err := decodeAttr(request.SearchAttributes)
	if err != nil {
		return err
	}

	msg := createVisibilityMessage(
		request.DomainUUID,
		request.WorkflowID,
		request.RunID,
		request.WorkflowTypeName,
		request.TaskList,
		request.StartTimestamp.UnixMilli(),
		request.ExecutionTimestamp.UnixMilli(),
		request.TaskID,
		request.Memo.Data,
		request.Memo.GetEncoding(),
		request.IsCron,
		request.NumClusters,
		string(attr),
		common.UpsertSearchAttributes,
		0, // will not be used
		0, // will not be used
		0, // will not be used
		request.UpdateTimestamp.UnixMilli(),
		request.ShardID,
	)
	return v.producer.Publish(ctx, msg)
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsQuery(request, false)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		v.logger.Fatal(fmt.Sprintf("AABBCCDDDEBUG%s, %s", request, rec))
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsQuery(request, true)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByType(ctx context.Context, request *p.InternalListWorkflowExecutionsByTypeRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsByTypeQuery(request, false)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByType(ctx context.Context, request *p.InternalListWorkflowExecutionsByTypeRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsByTypeQuery(request, true)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByWorkflowID(ctx context.Context, request *p.InternalListWorkflowExecutionsByWorkflowIDRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsByWorkflowIDQuery(request, false)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByWorkflowID(ctx context.Context, request *p.InternalListWorkflowExecutionsByWorkflowIDRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsByWorkflowIDQuery(request, true)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByStatus(ctx context.Context, request *p.InternalListClosedWorkflowExecutionsByStatusRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	query := getListWorkflowExecutionsByStatusQuery(request)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          isRecordValid,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) GetClosedWorkflowExecution(ctx context.Context, request *p.InternalGetClosedWorkflowExecutionRequest) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	query := getGetClosedWorkflowExecutionQuery(request)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          nil,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	resp, err := v.pinotClient.Search(req)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Pinot GetClosedWorkflowExecution failed, %v", err),
		}
	}

	return &p.InternalGetClosedWorkflowExecutionResponse{
		Execution: resp.Executions[0],
	}, nil
}

func (v *pinotVisibilityStore) ListWorkflowExecutions(ctx context.Context, request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	checkPageSize(request)

	// TODO: need to check next page token in the future

	query := getListWorkflowExecutionsByQueryQuery(request)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          nil,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) ScanWorkflowExecutions(ctx context.Context, request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	checkPageSize(request)

	// TODO: need to check next page token in the future

	query := getListWorkflowExecutionsByQueryQuery(request)

	req := &pnt.SearchRequest{
		Query:           query,
		IsOpen:          true,
		Filter:          nil,
		MaxResultWindow: v.config.ESIndexMaxResultWindow(),
	}

	return v.pinotClient.Search(req)
}

func (v *pinotVisibilityStore) CountWorkflowExecutions(ctx context.Context, request *p.CountWorkflowExecutionsRequest) (*p.CountWorkflowExecutionsResponse, error) {
	query := getCountWorkflowExecutionsQuery(request)
	resp, err := v.pinotClient.CountByQuery(query)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return &p.CountWorkflowExecutionsResponse{
		Count: resp,
	}, nil
}

func (v *pinotVisibilityStore) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) DeleteUninitializedWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	// temporary: not implemented, only implemented for ES
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) checkProducer() {
	if v.producer == nil {
		// must be bug, check history setup
		panic("message producer is nil")
	}
}

func createVisibilityMessage(
	// common parameters
	domainID string,
	wid,
	rid string,
	workflowTypeName string,
	taskList string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	taskID int64,
	memo []byte,
	encoding common.EncodingType,
	isCron bool,
	NumClusters int16,
	searchAttributes string,
	visibilityOperation common.VisibilityOperation,
	// specific to certain status
	closeTimeUnixNano int64, // close execution
	closeStatus workflow.WorkflowExecutionCloseStatus, // close execution
	historyLength int64, // close execution
	updateTimeUnixNano int64, // update execution,
	shardID int64,
) *indexer.PinotMessage {
	status := int(closeStatus)

	rawMsg := visibilityMessage{
		DocID:         wid + "-" + rid,
		DomainID:      domainID,
		WorkflowID:    wid,
		RunID:         rid,
		WorkflowType:  workflowTypeName,
		TaskList:      taskList,
		StartTime:     startTimeUnixNano,
		ExecutionTime: executionTimeUnixNano,
		IsCron:        isCron,
		NumClusters:   NumClusters,
		Attr:          searchAttributes,
		CloseTime:     closeTimeUnixNano,
		CloseStatus:   status,
		HistoryLength: historyLength,
		UpdateTime:    updateTimeUnixNano,
		ShardID:       shardID,
	}

	serializedMsg, err := json.Marshal(rawMsg)
	if err != nil {
		panic("serialize msg error!")
	}

	msgType := indexer.MessageTypePinot
	msg := &indexer.PinotMessage{
		MessageType: &msgType,
		WorkflowID:  common.StringPtr(wid),
		Payload:     &serializedMsg,
	}
	return msg
}

/****************************** Request Translator ******************************/

type PinotQuery struct {
	query   string
	filters PinotQueryFilter
	sorters string
	limits  string
}

func NewPinotQuery() PinotQuery {
	return PinotQuery{
		query:   fmt.Sprintf("SELECT *\nFROM %s\n", tableName),
		filters: PinotQueryFilter{},
		sorters: "",
		limits:  "",
	}
}

func NewPinotCountQuery() PinotQuery {
	return PinotQuery{
		query:   fmt.Sprintf("SELECT COUNT(*)\nFROM %s\n", tableName),
		filters: PinotQueryFilter{},
		sorters: "",
		limits:  "",
	}
}

func (q *PinotQuery) String() string {
	return fmt.Sprintf("%s%s%s%s", q.query, q.filters.string, q.sorters, q.limits)
}

func (q *PinotQuery) concatSorter(sorter string) {
	q.sorters += sorter + "\n"
}

func (q *PinotQuery) addPinotSorter(orderBy string, order string) {
	if q.sorters == "" {
		q.sorters = "Order BY "
	} else {
		q.sorters += ", "
	}
	q.sorters += fmt.Sprintf("%s %s\n", orderBy, order)
}

func (q *PinotQuery) addLimits(limit int) {
	q.limits += fmt.Sprintf("LIMIT %d\n", limit)
}

type PinotQueryFilter struct {
	string
}

func (f *PinotQueryFilter) checkFirstFilter() {
	if f.string == "" {
		f.string = "WHERE "
	} else {
		f.string += "AND "
	}
}

func (f *PinotQueryFilter) addEqual(obj string, val interface{}) {
	f.checkFirstFilter()
	quotedVal := fmt.Sprintf("'%s'", val)
	f.string += fmt.Sprintf("%s = %s\n", obj, quotedVal)
}

// addQuery adds a complete query into the filter
func (f *PinotQueryFilter) addQuery(query string) {
	f.checkFirstFilter()
	f.string += fmt.Sprintf("%s\n", query)
}

// addGte check object is greater than or equals to val
func (f *PinotQueryFilter) addGte(obj string, val interface{}) {
	f.checkFirstFilter()
	f.string += fmt.Sprintf("%s >= %s\n", obj, val)
}

// addLte check object is less than val
func (f *PinotQueryFilter) addLt(obj string, val interface{}) {
	f.checkFirstFilter()
	f.string += fmt.Sprintf("%s < %s\n", obj, val)
}

func (f *PinotQueryFilter) addTimeRange(obj string, earliest interface{}, latest interface{}) {
	f.checkFirstFilter()
	f.string += fmt.Sprintf("%s BETWEEN %v AND %v\n", obj, earliest, latest)
}

func getCountWorkflowExecutionsQuery(request *p.CountWorkflowExecutionsRequest) string {
	if request == nil {
		return ""
	}

	query := NewPinotCountQuery()

	// need to add Domain ID
	query.filters.addEqual(DomainID, request.DomainUUID)

	requestQuery := strings.TrimSpace(request.Query)
	if requestQuery != "" {
		query.filters.addQuery(request.Query)
	}

	return query.String()
}

func getListWorkflowExecutionsByQueryQuery(request *p.ListWorkflowExecutionsByQueryRequest) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	// need to add Domain ID
	query.filters.addEqual(DomainID, request.DomainUUID)

	requestQuery := strings.TrimSpace(request.Query)
	if requestQuery == "" {
		query.addLimits(request.PageSize)
	} else if common.IsJustOrderByClause(requestQuery) {
		query.concatSorter(requestQuery)
		query.addLimits(request.PageSize)
	} else {
		query.filters.addQuery(request.Query)
		query.addLimits(request.PageSize)
	}

	return query.String()
}

func getListWorkflowExecutionsQuery(request *p.InternalListWorkflowExecutionsRequest, isClosed bool) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	query.filters.addEqual(DomainID, request.DomainUUID)
	query.filters.addTimeRange(CloseTime, request.EarliestTime.UnixMilli(), request.LatestTime.UnixMilli()) //convert Unix Time to miliseconds
	if isClosed {
		query.filters.addGte(CloseStatus, "0")
	} else {
		query.filters.addLt(CloseStatus, "0")
	}

	query.addPinotSorter(CloseTime, DescendingOrder)
	query.addPinotSorter(RunID, DescendingOrder)
	return query.String()
}

func getListWorkflowExecutionsByTypeQuery(request *p.InternalListWorkflowExecutionsByTypeRequest, isClosed bool) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	query.filters.addEqual(DomainID, request.DomainUUID)
	query.filters.addEqual(WorkflowType, request.WorkflowTypeName)
	query.filters.addTimeRange(CloseTime, request.EarliestTime.UnixMilli(), request.LatestTime.UnixMilli()) //convert Unix Time to miliseconds
	if isClosed {
		query.filters.addGte(CloseStatus, "0")
	} else {
		query.filters.addLt(CloseStatus, "0")
	}

	query.addPinotSorter(CloseTime, DescendingOrder)
	query.addPinotSorter(RunID, DescendingOrder)
	return query.String()
}

func getListWorkflowExecutionsByWorkflowIDQuery(request *p.InternalListWorkflowExecutionsByWorkflowIDRequest, isClosed bool) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	query.filters.addEqual(DomainID, request.DomainUUID)
	query.filters.addEqual(WorkflowID, request.WorkflowID)
	query.filters.addTimeRange(CloseTime, request.EarliestTime.UnixMilli(), request.LatestTime.UnixMilli()) //convert Unix Time to miliseconds
	if isClosed {
		query.filters.addGte(CloseStatus, "0")
	} else {
		query.filters.addLt(CloseStatus, "0")
	}

	query.addPinotSorter(CloseTime, DescendingOrder)
	query.addPinotSorter(RunID, DescendingOrder)
	return query.String()
}

func getListWorkflowExecutionsByStatusQuery(request *p.InternalListClosedWorkflowExecutionsByStatusRequest) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	query.filters.addEqual(DomainID, request.DomainUUID)

	status := "0"
	switch request.Status.String() {
	case "COMPLETED":
		status = "0"
	case "FAILED":
		status = "1"
	case "CANCELED":
		status = "2"
	case "TERMINATED":
		status = "3"
	case "CONTINUED_AS_NEW":
		status = "4"
	case "TIMED_OUT":
		status = "5"
	}

	query.filters.addEqual(CloseStatus, status)
	query.filters.addTimeRange(CloseTime, request.EarliestTime.UnixMilli(), request.LatestTime.UnixMilli()) //convert Unix Time to miliseconds

	query.addPinotSorter(CloseTime, DescendingOrder)
	query.addPinotSorter(RunID, DescendingOrder)
	return query.String()
}

func getGetClosedWorkflowExecutionQuery(request *p.InternalGetClosedWorkflowExecutionRequest) string {
	if request == nil {
		return ""
	}

	query := NewPinotQuery()

	query.filters.addEqual(DomainID, request.DomainUUID)
	query.filters.addGte(CloseStatus, "0")
	query.filters.addEqual(WorkflowID, request.Execution.GetWorkflowID())

	rid := request.Execution.GetRunID()
	if rid != "" {
		query.filters.addEqual(RunID, rid)
	}

	return query.String()
}

func checkPageSize(request *p.ListWorkflowExecutionsByQueryRequest) {
	if request.PageSize == 0 {
		request.PageSize = 1000
	}
}
