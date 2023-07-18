package persistence

import (
	"bytes"
	"context"
	"fmt"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
	"reflect"
	"strings"
)

func interfaceToMap(in interface{}) map[string][]byte {
	if in == nil || in == "" {
		return map[string][]byte{}
	}

	v, ok := in.(map[string][]byte)
	if !ok {
		panic(fmt.Sprintf("interface to map error in ES/Pinot comparator: %#v", in))
	}

	return v
}

func compareSearchAttributes(esSearchAttribute interface{}, pinotSearchAttribute interface{}) error {
	esAttr, ok := esSearchAttribute.(*types.SearchAttributes)
	if !ok {
		panic("interface is not a SearchAttributes! ")
	}

	pinotAttr, ok := pinotSearchAttribute.(*types.SearchAttributes)
	if !ok {
		panic("interface is not a SearchAttributes! ")
	}

	esSearchAttributeList := interfaceToMap(esAttr.GetIndexedFields())
	pinotSearchAttributeList := interfaceToMap(pinotAttr.GetIndexedFields())
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
		panic("interface is not a WorkflowExecution! ")
	}

	pinotExecution, ok := pinotInput.(*types.WorkflowExecution)
	if !ok {
		panic("interface is not a WorkflowExecution! ")
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
		panic("interface is not a WorkflowType! ")
	}

	pinotType, ok := pinotInput.(*types.WorkflowType)
	if !ok {
		panic("interface is not a WorkflowType! ")
	}

	if esType.GetName() != pinotType.GetName() {
		return fmt.Errorf(fmt.Sprintf("Comparison Failed: WorkflowTypes are not equal. ES value = %s, Pinot value = %s", esType.GetName(), pinotType.GetName()))
	}

	return nil
}

func compareCloseStatus(esInput interface{}, pinotInput interface{}) error {
	esStatus, ok := esInput.(*types.WorkflowExecutionCloseStatus)
	if !ok {
		panic("interface is not a WorkflowExecutionCloseStatus! ")
	}

	pinotStatus, ok := pinotInput.(*types.WorkflowExecutionCloseStatus)
	if !ok {
		panic("interface is not a WorkflowExecutionCloseStatus! ")
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

func comparePinotESListResponse(
	ESManager VisibilityManager,
	PinotManager VisibilityManager,
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
	logger log.Logger,
) (*ListWorkflowExecutionsResponse, error) {
	esResponse, err := ESManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		logger.Error(fmt.Sprintf("ListOpenWorkflowExecutions in comparator, ES: %s", err))
	}

	pinotResponse, err := PinotManager.ListOpenWorkflowExecutions(ctx, request)
	if err != nil {
		logger.Error(fmt.Sprintf("ListOpenWorkflowExecutions in comparator, Pinot: %s", err))
	}

	err = compareListWorkflowExecutions(esResponse.Executions, pinotResponse.Executions)
	if err != nil {
		return nil, err
	}
	return esResponse, nil
}
