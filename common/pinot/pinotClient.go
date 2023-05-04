// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pinot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/startreedata/pinot-client-go/pinot"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

type PinotClient struct {
	client    *pinot.Connection
	logger    log.Logger
	tableName string
}

func NewPinotClient(client *pinot.Connection, logger log.Logger, tableName string) GenericClient {
	return &PinotClient{
		client:    client,
		logger:    logger,
		tableName: tableName,
	}
}

func (c *PinotClient) Search(request *SearchRequest) (*SearchResponse, error) {
	resp, err := c.client.ExecuteSQL(c.tableName, request.Query)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Pinot Search failed, %v", err),
		}
	}

	token, err := GetNextPageToken(request.ListRequest.NextPageToken)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Get NextPage token failed, %v", err),
		}
	}

	return c.getInternalListWorkflowExecutionsResponse(resp, request.Filter, token, request.ListRequest.PageSize, request.MaxResultWindow)
}

func (c *PinotClient) CountByQuery(query string) (int64, error) {
	resp, err := c.client.ExecuteSQL(c.tableName, query)
	if err != nil {
		return 0, &types.InternalServiceError{
			Message: fmt.Sprintf("CountWorkflowExecutions ExecuteSQL failed, %v", err),
		}
	}

	count, err := resp.ResultTable.Rows[0][0].(json.Number).Int64()
	if err == nil {
		return count, nil
	}

	return -1, &types.InternalServiceError{
		Message: fmt.Sprintf("can't convert result to integer!, query = %s, query result = %v, err = %v", query, resp.ResultTable.Rows[0][0], err),
	}
}

func (c *PinotClient) GetTableName() string {
	return c.tableName
}

/****************************** Response Translator ******************************/

func buildMap(hit []interface{}, columnNames []string) map[string]interface{} {
	resMap := make(map[string]interface{})

	for i := 0; i < len(columnNames); i++ {
		key := columnNames[i]
		// checks if it is system key, if yes, put it into the map
		if ok, _ := isSystemKey(key); ok {
			resMap[key] = hit[i]
		}
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
	Attr          string
}

func (c *PinotClient) convertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string) *p.InternalVisibilityWorkflowExecutionInfo {
	if len(hit) != len(columnNames) {
		return nil
	}

	columnNameToValue := buildMap(hit, columnNames)
	jsonColumnNameToValue, err := json.Marshal(columnNameToValue)
	if err != nil { // log and skip error
		c.logger.Error("unable to marshal columnNameToValue",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	var source *VisibilityRecord
	err = json.Unmarshal(jsonColumnNameToValue, &source)
	if err != nil { // log and skip error
		c.logger.Error("unable to Unmarshal columnNameToValue",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	attr := make(map[string]interface{})
	if source.Attr != "null" {
		err = json.Unmarshal([]byte(source.Attr), &attr)

		if err != nil { // log and skip error
			c.logger.Error("unable to Unmarshal source.Attr",
				tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
			)
			return nil
		}
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
		SearchAttributes: attr,
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

func (c *PinotClient) getInternalListWorkflowExecutionsResponse(
	resp *pinot.BrokerResponse,
	isRecordValid func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool,
	token *PinotVisibilityPageToken,
	pageSize int,
	maxResultWindow int,
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
		workflowExecutionInfo := c.convertSearchResultToVisibilityRecord(actualHits[i], columnNames)
		if isRecordValid == nil || isRecordValid(workflowExecutionInfo) {
			response.Executions = append(response.Executions, workflowExecutionInfo)
		}
	}

	if numOfActualHits == pageSize { // this means the response is not the last page
		var nextPageToken []byte
		var err error

		// ES Search API support pagination using From and PageSize, but has limit that From+PageSize cannot exceed a threshold
		// TODO: need to confirm if pinot has similar settings
		// don't need to retrieve deeper pages in pinot, and no functions like ES SearchAfter
		if resp.NumDocsScanned <= int64(maxResultWindow-pageSize) { // use pinot Search From+Size
			nextPageToken, err = SerializePageToken(&PinotVisibilityPageToken{From: token.From + numOfActualHits})
		}
		if err != nil {
			return nil, err
		}

		response.NextPageToken = make([]byte, len(nextPageToken))
		copy(response.NextPageToken, nextPageToken)
	}

	return response, nil
}

func (c *PinotClient) getInternalGetClosedWorkflowExecutionResponse(resp *pinot.BrokerResponse) (
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
	response.Execution = c.convertSearchResultToVisibilityRecord(actualHits[0], columnNames)

	return response, nil
}

// checks if a string is system key
func isSystemKey(key string) (bool, string) {
	msg := VisibilityRecord{}
	values := reflect.ValueOf(msg)
	typesOf := values.Type()
	for i := 0; i < values.NumField(); i++ {
		fieldName := typesOf.Field(i).Name
		if fieldName == key {
			return true, typesOf.Field(i).Type.String()
		}
	}
	return false, "nil"
}
