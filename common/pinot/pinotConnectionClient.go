package pinot

import (
	"encoding/json"
	"fmt"
	"github.com/startreedata/pinot-client-go/pinot"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
	"time"
)

const (
	tableName = "cadence_visibility_pinot"
)

type pinotConnectionClient struct {
	client *pinot.Connection
	logger log.Logger
}

func NewPinotConnectionClient(client *pinot.Connection, logger log.Logger) *pinotConnectionClient {
	return &pinotConnectionClient{
		client: client,
		logger: logger,
	}
}

func (c *pinotConnectionClient) Search(request *SearchRequest) (*SearchResponse, error) {
	resp, err := c.client.ExecuteSQL(tableName, request.Query)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Pinot Search failed, %v", err),
		}
	}

	return c.getInternalListWorkflowExecutionsResponse(resp, request.Filter)
}

func (c *pinotConnectionClient) CountByQuery(query string) (int64, error) {
	resp, err := c.client.ExecuteSQL(tableName, query)
	if err != nil {
		return 0, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return int64(resp.ResultTable.GetRowCount()), nil
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

func (c *pinotConnectionClient) convertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string) *p.InternalVisibilityWorkflowExecutionInfo {
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
		c.logger.Error("unable to marshal columnNameToValue",
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

func (c *pinotConnectionClient) getInternalListWorkflowExecutionsResponse(
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
		workflowExecutionInfo := c.convertSearchResultToVisibilityRecord(actualHits[i], columnNames)
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

func (c *pinotConnectionClient) getInternalGetClosedWorkflowExecutionResponse(resp *pinot.BrokerResponse) (
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
