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

package persistence

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/uber/cadence/common/log/tag"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

func interfaceToMap(in interface{}) (map[string][]byte, error) {
	if in == nil || in == "" {
		return map[string][]byte{}, nil
	}

	v, ok := in.(map[string][]byte)
	if !ok {
		return map[string][]byte{}, fmt.Errorf(fmt.Sprintf("interface to map error in ES/Pinot comparator: %#v", in))
	}

	return v, nil
}

func compareSearchAttributes(esSearchAttribute interface{}, pinotSearchAttribute interface{}) error {
	esAttr, ok := esSearchAttribute.(*types.SearchAttributes)
	if !ok {
		return fmt.Errorf("interface is not an ES SearchAttributes! ")
	}

	pinotAttr, ok := pinotSearchAttribute.(*types.SearchAttributes)
	if !ok {
		return fmt.Errorf("interface is not a pinot SearchAttributes! ")
	}

	esSearchAttributeList, err := interfaceToMap(esAttr.GetIndexedFields())
	if err != nil {
		return err
	}
	pinotSearchAttributeList, err := interfaceToMap(pinotAttr.GetIndexedFields())
	if err != nil {
		return err
	}

	for key, esValue := range esSearchAttributeList { // length(esAttribute) <= length(pinotAttribute)
		pinotValue := pinotSearchAttributeList[key]
		if !bytes.Equal(esValue, pinotValue) {
			return fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = %s, Pinot value = %s", key, esValue, pinotValue))
		}
	}

	return nil
}

func compareExecutions(esInput interface{}, pinotInput interface{}) error {
	esExecution, ok := esInput.(*types.WorkflowExecution)
	if !ok {
		return fmt.Errorf("interface is not an ES WorkflowExecution! ")
	}

	pinotExecution, ok := pinotInput.(*types.WorkflowExecution)
	if !ok {
		return fmt.Errorf("interface is not a pinot WorkflowExecution! ")
	}

	if esExecution.GetWorkflowID() != pinotExecution.GetWorkflowID() {
		return fmt.Errorf(fmt.Sprintf("Comparison Failed: Execution.WorkflowID are not equal. ES value = %s, Pinot value = %s", esExecution.GetWorkflowID(), pinotExecution.GetWorkflowID()))
	}

	if esExecution.GetRunID() != pinotExecution.GetRunID() {
		return fmt.Errorf(fmt.Sprintf("Comparison Failed: Execution.RunID are not equal. ES value = %s, Pinot value = %s", esExecution.GetRunID(), pinotExecution.GetRunID()))
	}

	return nil
}

func compareType(esInput interface{}, pinotInput interface{}) error {
	esType, ok := esInput.(*types.WorkflowType)
	if !ok {
		return fmt.Errorf("interface is not an ES WorkflowType! ")
	}

	pinotType, ok := pinotInput.(*types.WorkflowType)
	if !ok {
		return fmt.Errorf("interface is not a pinot WorkflowType! ")
	}

	if esType.GetName() != pinotType.GetName() {
		return fmt.Errorf(fmt.Sprintf("Comparison Failed: WorkflowTypes are not equal. ES value = %s, Pinot value = %s", esType.GetName(), pinotType.GetName()))
	}

	return nil
}

func compareCloseStatus(esInput interface{}, pinotInput interface{}) error {
	esStatus, ok := esInput.(*types.WorkflowExecutionCloseStatus)
	if !ok {
		return fmt.Errorf("interface is not an ES WorkflowExecutionCloseStatus! ")
	}

	pinotStatus, ok := pinotInput.(*types.WorkflowExecutionCloseStatus)
	if !ok {
		return fmt.Errorf("interface is not a pinot WorkflowExecutionCloseStatus! ")
	}

	if esStatus != pinotStatus {
		return fmt.Errorf(fmt.Sprintf("Comparison Failed: WorkflowExecutionCloseStatus are not equal. ES value = %s, Pinot value = %s", esStatus, pinotStatus))
	}

	return nil
}

func compareListWorkflowExecutionInfo(
	esExecutionInfo *types.WorkflowExecutionInfo,
	pinotExecutionInfo *types.WorkflowExecutionInfo,
) error {
	vOfES := reflect.ValueOf(*esExecutionInfo)
	typeOfesExecutionInfo := vOfES.Type()
	vOfPinot := reflect.ValueOf(*pinotExecutionInfo)

	for i := 0; i < vOfES.NumField(); i++ {
		esFieldName := typeOfesExecutionInfo.Field(i).Name
		esValue := vOfES.Field(i).Interface()
		pinotValue := vOfPinot.Field(i).Interface()

		// if the value in ES is nil, then we don't need to compare
		if esValue == nil {
			continue
		}

		// if the value in ES is not nil but in pinot is nil, then there's an error
		if pinotValue == nil {
			return fmt.Errorf("Pinot result is nil while ES result is not. ")
		}

		switch strings.ToLower(esFieldName) {
		case "memo", "autoresetpoints", "partitionconfig":

		case "searchattributes":
			err := compareSearchAttributes(esValue, pinotValue)
			if err != nil {
				return err
			}
		case "execution", "parentexecution":
			err := compareExecutions(esValue, pinotValue)
			if err != nil {
				return err
			}
		case "type":
			err := compareType(esValue, pinotValue)
			if err != nil {
				return err
			}
		case "closestatus":
			err := compareCloseStatus(esValue, pinotValue)
			if err != nil {
				return err
			}
		default:
			if esValue != pinotValue {
				return fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = %s, Pinot value = %s", esFieldName, esValue, pinotValue))
			}
		}
	}

	return nil
}

func compareListWorkflowExecutions(
	esExecutionInfos []*types.WorkflowExecutionInfo,
	pinotExecutionInfos []*types.WorkflowExecutionInfo,
) error {
	if esExecutionInfos == nil && pinotExecutionInfos == nil {
		return nil
	}
	if esExecutionInfos == nil || pinotExecutionInfos == nil {
		return fmt.Errorf(fmt.Sprintf("Comparison failed. One of the response is nil. "))
	}
	if len(esExecutionInfos) != len(pinotExecutionInfos) {
		return fmt.Errorf(fmt.Sprintf("Comparison failed. result length doesn't equal. "))
	}

	for i := 0; i < len(esExecutionInfos); i++ {
		err := compareListWorkflowExecutionInfo(esExecutionInfos[i], pinotExecutionInfos[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func comparePinotESListOpenResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutions in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf(fmt.Sprintf("ListOpenWorkflowExecutions in comparator error, Pinot: %s", err)))

	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListClosedResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutions in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutions in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListOpenByTypeResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByTypeRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListOpenWorkflowExecutionsByType(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutionsByType in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListOpenWorkflowExecutionsByType(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutionsByType in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListClosedByTypeResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByTypeRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListClosedWorkflowExecutionsByType(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByType in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListClosedWorkflowExecutionsByType(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByType in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListOpenByWorkflowIDResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutionsByWorkflowID in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutionsByWorkflowID in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListClosedByWorkflowIDResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByWorkflowID in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByWorkflowID in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListClosedByStatusResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListClosedWorkflowExecutionsByStatusRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListClosedWorkflowExecutionsByStatus(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByStatus in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListClosedWorkflowExecutionsByStatus(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListClosedWorkflowExecutionsByStatus in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESGetClosedByStatusResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *GetClosedWorkflowExecutionRequest,
	logger log.Logger,
) (*GetClosedWorkflowExecutionResponse, error) {
	esResponse, err := ESManager.GetClosedWorkflowExecution(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("GetClosedWorkflowExecutions in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.GetClosedWorkflowExecution(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("GetClosedWorkflowExecutions in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutionInfo(esResponse.Execution, pinotResponse.Execution)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESListByQueryResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByQueryRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListWorkflowExecutionsByQuery in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ListOpenWorkflowExecutionsByQuery in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESScanResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *ListWorkflowExecutionsByQueryRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ScanWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ScanWorkflowExecutions in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ScanWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("ScanWorkflowExecutions in comparator error, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}
	return esResponse, nil
}

func comparePinotESCountResponse(
	ctx context.Context,
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	request *CountWorkflowExecutionsRequest,
	logger log.Logger,
) (*CountWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.CountWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("CountOpenWorkflowExecutions in comparator error, ES: %s", err))
	}

	pinotResponse, err := PinotManager.CountWorkflowExecutions(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("CountOpenWorkflowExecutions in comparator error, Pinot: %s", err))
	}

	if esResponse.Count != pinotResponse.Count {
		err = fmt.Errorf(fmt.Sprintf("Comparison Failed: counts are not equal. ES value = %v, Pinot value = %v", esResponse.Count, pinotResponse.Count))
		logger.Error("ES/Pinot Response comparison Error! ", tag.Error(err))
		return esResponse, nil
	}

	return esResponse, nil
}
