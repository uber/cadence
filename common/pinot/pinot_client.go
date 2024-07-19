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

	"github.com/startreedata/pinot-client-go/pinot"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type PinotClient struct {
	client      *pinot.Connection
	logger      log.Logger
	tableName   string
	serviceName string
}

func NewPinotClient(client *pinot.Connection, logger log.Logger, pinotConfig *config.PinotVisibilityConfig) GenericClient {
	return &PinotClient{
		client:      client,
		logger:      logger,
		tableName:   pinotConfig.Table,
		serviceName: pinotConfig.ServiceName,
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

func (c *PinotClient) SearchAggr(request *SearchRequest) (AggrResponse, error) {
	resp, err := c.client.ExecuteSQL(c.tableName, request.Query)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Pinot SearchAggr failed, %v", err),
		}
	}

	return resp.ResultTable.Rows, nil
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

// Pinot Response Translator
// We flattened the search attributes into columns in Pinot table
// This function converts the search result back to VisibilityRecord
func (c *PinotClient) getInternalListWorkflowExecutionsResponse(
	resp *pinot.BrokerResponse,
	isRecordValid func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool,
	token *PinotVisibilityPageToken,
	pageSize int,
	maxResultWindow int,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	response := &p.InternalListWorkflowExecutionsResponse{}
	if resp == nil || resp.ResultTable == nil || resp.ResultTable.GetRowCount() == 0 {
		return response, nil
	}
	schema := resp.ResultTable.DataSchema // get the schema to map results
	columnNames := schema.ColumnNames
	actualHits := resp.ResultTable.Rows
	numOfActualHits := resp.ResultTable.GetRowCount()
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo, err := ConvertSearchResultToVisibilityRecord(actualHits[i], columnNames)
		if err != nil {
			return nil, err
		}

		if isRecordValid == nil || isRecordValid(workflowExecutionInfo) {
			response.Executions = append(response.Executions, workflowExecutionInfo)
		}
	}

	if numOfActualHits == pageSize { // this means the response is not the last page
		var nextPageToken []byte
		var err error

		// ES Search API support pagination using From and PageSize, but has limit that From+PageSize cannot exceed a threshold
		// In pinot we just skip (previous pages * page limit) items and take the next (number of page limit) items
		nextPageToken, err = SerializePageToken(&PinotVisibilityPageToken{From: token.From + numOfActualHits})

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
	var err error
	response.Execution, err = ConvertSearchResultToVisibilityRecord(actualHits[0], columnNames)
	if err != nil {
		return nil, err
	}

	return response, nil
}
