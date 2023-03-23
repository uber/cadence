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
	"time"

	"github.com/startreedata/pinot-client-go/pinot"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

const (
	pinotPersistenceName = "pinot"
	tableName            = "cadence-visibility-pinot"
	DescendingOrder      = "DESC"
	AcendingOrder        = "ASC"

	DomainID            = "DomainID"
	WorkflowID          = "WorkflowID"
	RunID               = "RunID"
	WorkflowType        = "WorkflowType"
	StartTime           = "StartTime"
	ExecutionTime       = "ExecutionTime"
	CloseTime           = "CloseTime"
	CloseStatus         = "CloseStatus"
	HistoryLength       = "HistoryLength"
	Memo                = "Memo"
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
		pinotClient *pinot.Connection
		producer    messaging.Producer
		logger      log.Logger
		config      *service.Config
	}

	visibilityMessage struct {
		DomainID              string                     `json:"domainID,omitempty"`
		Wid                   string                     `json:"wid,omitempty"`
		Rid                   string                     `json:"rid,omitempty"`
		WorkflowTypeName      string                     `json:"workflowTypeName,omitempty"`
		TaskList              string                     `json:"taskList,omitempty"`
		StartTimeUnixNano     int64                      `json:"startTimeUnixNano,omitempty"`
		ExecutionTimeUnixNano int64                      `json:"executionTimeUnixNano,omitempty"`
		TaskID                int64                      `json:"taskID,omitempty"`
		Memo                  []byte                     `json:"memo,omitempty"`
		Encoding              common.EncodingType        `json:"encoding,omitempty"`
		IsCron                bool                       `json:"isCron,omitempty"`
		NumClusters           int16                      `json:"numClusters,omitempty"`
		SearchAttributes      map[string][]byte          `json:"searchAttributes,omitempty"`
		VisibilityOperation   common.VisibilityOperation `json:"visibilityOperation,omitempty"`
		// specific to certain status
		EndTimeUnixNano    int64                                 `json:"endTimeUnixNano,omitempty"`    // close execution
		CloseStatus        workflow.WorkflowExecutionCloseStatus `json:"closeStatus,omitempty"`        // close execution
		HistoryLength      int64                                 `json:"historyLength,omitempty"`      // close execution
		UpdateTimeUnixNano int64                                 `json:"updateTimeUnixNano,omitempty"` // update execution,
		ShardID            int64                                 `json:"shardID,omitempty"`
	}
)

var _ p.VisibilityStore = (*pinotVisibilityStore)(nil)

func NewPinotVisibilityStore(
	pinotClient *pinot.Connection,
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
	v.pinotClient.CloseTrace() // TODO: need to double check what is close trace do. Does it close the client?
}

func (v *pinotVisibilityStore) GetName() string {
	return pinotPersistenceName
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	v.checkProducer()
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
		request.SearchAttributes,
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
		request.SearchAttributes,
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
		nil,
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
		request.SearchAttributes,
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

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsQuery(request, false)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsQuery(request, true)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByType(ctx context.Context, request *p.InternalListWorkflowExecutionsByTypeRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsByTypeQuery(request, false)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByType(ctx context.Context, request *p.InternalListWorkflowExecutionsByTypeRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsByTypeQuery(request, true)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByWorkflowID(ctx context.Context, request *p.InternalListWorkflowExecutionsByWorkflowIDRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsByWorkflowIDQuery(request, false)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByWorkflowID(ctx context.Context, request *p.InternalListWorkflowExecutionsByWorkflowIDRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsByWorkflowIDQuery(request, true)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByStatus(ctx context.Context, request *p.InternalListClosedWorkflowExecutionsByStatusRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return !request.EarliestTime.After(rec.CloseTime) && !rec.CloseTime.After(request.LatestTime)
	}

	ListClosedWorkflowExecutionsQuery := getListWorkflowExecutionsByStatusQuery(request)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalListWorkflowExecutionsResponse(resp, isRecordValid)
}

func (v *pinotVisibilityStore) GetClosedWorkflowExecution(ctx context.Context, request *p.InternalGetClosedWorkflowExecutionRequest) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	ListClosedWorkflowExecutionsQuery := getGetClosedWorkflowExecutionQuery(request)
	resp, err := v.pinotClient.ExecuteSQL(tableName, ListClosedWorkflowExecutionsQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}
	return v.getInternalGetClosedWorkflowExecutionResponse(resp)
}

func (v *pinotVisibilityStore) ListWorkflowExecutions(ctx context.Context, request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	checkPageSize(request)

	// TODO: need to check next page token in the future

	workflowExecutionQuery := getListWorkflowExecutionsByQueryQuery(request)
	resp, err := v.pinotClient.ExecuteSQL(tableName, workflowExecutionQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return v.getInternalListWorkflowExecutionsResponse(resp, nil)
}

func (v *pinotVisibilityStore) ScanWorkflowExecutions(ctx context.Context, request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	checkPageSize(request)

	// TODO: need to check next page token in the future

	workflowExecutionQuery := getListWorkflowExecutionsByQueryQuery(request)
	resp, err := v.pinotClient.ExecuteSQL(tableName, workflowExecutionQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return v.getInternalListWorkflowExecutionsResponse(resp, nil)
}

func (v *pinotVisibilityStore) CountWorkflowExecutions(ctx context.Context, request *p.CountWorkflowExecutionsRequest) (*p.CountWorkflowExecutionsResponse, error) {
	workflowExecutionQuery := getCountWorkflowExecutionsQuery(request)
	resp, err := v.pinotClient.ExecuteSQL(tableName, workflowExecutionQuery)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return &p.CountWorkflowExecutionsResponse{
		Count: int64(resp.ResultTable.GetRowCount()),
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

func checkPageSize(request *p.ListWorkflowExecutionsByQueryRequest) {
	if request.PageSize == 0 {
		request.PageSize = 1000
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
	searchAttributes map[string][]byte,
	visibilityOperation common.VisibilityOperation,
	// specific to certain status
	endTimeUnixNano int64, // close execution
	closeStatus workflow.WorkflowExecutionCloseStatus, // close execution
	historyLength int64, // close execution
	updateTimeUnixNano int64, // update execution,
	shardID int64,
) []byte {
	msg := visibilityMessage{
		DomainID:              domainID,
		Wid:                   wid,
		Rid:                   rid,
		WorkflowTypeName:      workflowTypeName,
		TaskList:              taskList,
		StartTimeUnixNano:     startTimeUnixNano,
		ExecutionTimeUnixNano: executionTimeUnixNano,
		TaskID:                taskID,
		Memo:                  memo,
		Encoding:              encoding,
		IsCron:                isCron,
		NumClusters:           NumClusters,
		SearchAttributes:      searchAttributes,
		VisibilityOperation:   visibilityOperation,
		EndTimeUnixNano:       endTimeUnixNano,
		CloseStatus:           closeStatus,
		HistoryLength:         historyLength,
		UpdateTimeUnixNano:    updateTimeUnixNano,
		ShardID:               shardID,
	}

	serializedMsg, err := json.Marshal(msg)
	if err != nil {
		panic("serialize msg error!")
	}

	return serializedMsg
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
	q.sorters += fmt.Sprintf("Order BY %s %s\n", orderBy, order)
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
	f.string += fmt.Sprintf("%s = %s\n", obj, val)
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

/****************************** Response Translator ******************************/

func buildMap(hit []interface{}, columnNames []string) map[string]interface{} {
	resMap := make(map[string]interface{})

	for i := 0; i < len(columnNames); i++ {
		resMap[columnNames[i]] = hit[i]
	}

	return resMap
}

// VisibilityRecord is a struct of doc for deserialization
type VisibilityRecord struct {
	WorkflowID    string
	RunID         string
	WorkflowType  string
	DomainID      string
	StartTime     int64
	ExecutionTime int64
	CloseTime     int64
	CloseStatus   workflow.WorkflowExecutionCloseStatus
	HistoryLength int64
	Encoding      string
	TaskList      string
	IsCron        bool
	NumClusters   int16
	UpdateTime    int64
	Attr          map[string]interface{}
}

func (v *pinotVisibilityStore) convertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string) *p.InternalVisibilityWorkflowExecutionInfo {
	if len(hit) != len(columnNames) {
		return nil
	}

	columnNameToValue := buildMap(hit, columnNames)
	jsonColumnNameToValue, err := json.Marshal(columnNameToValue)
	if err != nil { // log and skip error
		v.logger.Error("unable to marshal columnNameToValue",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	var source *VisibilityRecord
	err = json.Unmarshal(jsonColumnNameToValue, &source)
	if err != nil { // log and skip error
		v.logger.Error("unable to marshal columnNameToValue",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	record := &p.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         source.DomainID,
		WorkflowType:     source.WorkflowType,
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.UnixMilli(source.StartTime), // be careful: source.StartTime is in milisecond
		ExecutionTime:    time.UnixMilli(source.ExecutionTime),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		NumClusters:      source.NumClusters,
		SearchAttributes: source.Attr,
	}
	if source.UpdateTime != 0 {
		record.UpdateTime = time.UnixMilli(source.UpdateTime)
	}
	if source.CloseTime != 0 {
		record.CloseTime = time.UnixMilli(source.CloseTime)
		record.Status = thrift.ToWorkflowExecutionCloseStatus(&source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func (v *pinotVisibilityStore) getInternalListWorkflowExecutionsResponse(
	resp *pinot.BrokerResponse,
	isRecordValid func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if resp == nil {
		return nil, nil
	}

	response := &p.InternalListWorkflowExecutionsResponse{}

	schema := resp.ResultTable.DataSchema // get the schema to map results
	//columnDataTypes := schema.ColumnDataTypes
	columnNames := schema.ColumnNames
	actualHits := resp.ResultTable.Rows

	numOfActualHits := resp.ResultTable.GetRowCount()

	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)

	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo := v.convertSearchResultToVisibilityRecord(actualHits[i], columnNames)
		if isRecordValid == nil || isRecordValid(workflowExecutionInfo) {
			response.Executions = append(response.Executions, workflowExecutionInfo)
		}
	}

	//if numOfActualHits == pageSize { // this means the response is not the last page
	//	var nextPageToken []byte
	//	var err error
	//
	//	// ES Search API support pagination using From and PageSize, but has limit that From+PageSize cannot exceed a threshold
	//	// to retrieve deeper pages, use ES SearchAfter
	//	if searchHits.TotalHits <= int64(maxResultWindow-pageSize) { // use ES Search From+Size
	//		nextPageToken, err = SerializePageToken(&ElasticVisibilityPageToken{From: token.From + numOfActualHits})
	//	} else { // use ES Search After
	//		var sortVal interface{}
	//		sortVals := actualHits[len(response.Executions)-1].Sort
	//		sortVal = sortVals[0]
	//		tieBreaker := sortVals[1].(string)
	//
	//		nextPageToken, err = SerializePageToken(&ElasticVisibilityPageToken{SortValue: sortVal, TieBreaker: tieBreaker})
	//	}
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	response.NextPageToken = make([]byte, len(nextPageToken))
	//	copy(response.NextPageToken, nextPageToken)
	//}

	return response, nil
}

func (v *pinotVisibilityStore) getInternalGetClosedWorkflowExecutionResponse(resp *pinot.BrokerResponse) (
	*p.InternalGetClosedWorkflowExecutionResponse,
	error,
) {
	if resp == nil {
		return nil, nil
	}

	response := &p.InternalGetClosedWorkflowExecutionResponse{}
	schema := resp.ResultTable.DataSchema // get the schema to map results
	columnNames := schema.ColumnNames
	actualHits := resp.ResultTable.Rows
	response.Execution = v.convertSearchResultToVisibilityRecord(actualHits[0], columnNames)

	return response, nil
}
