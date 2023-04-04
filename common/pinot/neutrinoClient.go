package pinot

import (
	"fmt"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type NeutrinoClient struct {
	client *Gateway
	logger log.Logger
}

func NewNeutrinoClient(client *Gateway, logger log.Logger) GenericClient {
	return &NeutrinoClient{
		client: client,
		logger: logger,
	}
}

func (c *NeutrinoClient) Search(request *SearchRequest) (*SearchResponse, error) {
	resp, err := c.client.Query(request.Query)

	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Pinot Search failed, %v", err),
		}
	}

	return c.getInternalListWorkflowExecutionsResponse(resp, request.Filter)
}

func (c *NeutrinoClient) CountByQuery(query string) (int64, error) {
	resp, err := c.client.Query(query)
	if err != nil {
		return 0, &types.InternalServiceError{
			Message: fmt.Sprintf("ListClosedWorkflowExecutions failed, %v", err),
		}
	}

	return int64(len(resp.Data)), nil
}

/****************************** Response Translator ******************************/

func (c *NeutrinoClient) getInternalListWorkflowExecutionsResponse(
	resp *queryResponse,
	isRecordValid func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if resp == nil {
		return nil, nil
	}

	response := &p.InternalListWorkflowExecutionsResponse{}

	//columnDataTypes := schema.ColumnDataTypes
	columns := resp.Columns

	columnNames := make([]string, len(columns))
	for i := 0; i < len(columns); i++ {
		columnNames[i] = columns[i].Name
	}

	actualHits := resp.Data

	numOfActualHits := len(actualHits)

	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)

	for i := 0; i < numOfActualHits; i++ {
		workflowExecutionInfo := convertSearchResultToVisibilityRecord(actualHits[i], columnNames)
		if isRecordValid == nil || isRecordValid(workflowExecutionInfo) {
			response.Executions = append(response.Executions, workflowExecutionInfo)
		}
	}

	return response, nil
}
