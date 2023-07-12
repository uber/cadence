package persistence

import (
	"context"
	"fmt"
	"github.com/uber/cadence/common/log"
	"reflect"
	"strings"
)

func interfaceToMap(in interface{}) map[string]interface{} {
	v, ok := in.(map[string]interface{})
	if !ok {
		panic("interface to map error in ES/Pinot comparator! ")
	}
	return v
}

func compareSearchAttributes(esSearchAttribute interface{}, pinotSearchAttribute interface{}) error {
	esSearchAttributeList := interfaceToMap(esSearchAttribute)
	pinotSearchAttributeList := interfaceToMap(pinotSearchAttribute)
	for key, esValue := range esSearchAttributeList {
		pinotValue := pinotSearchAttributeList[key]
		if pinotSearchAttributeList[key] != esValue {
			return fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = %s, Pinot value = %s", key, esValue, pinotValue))
		}
	}

	return nil
}

func compareListWorkflowExecutionsResponse(
	pinotResponse *ListWorkflowExecutionsResponse,
	esResponse *ListWorkflowExecutionsResponse,
) (*ListWorkflowExecutionsResponse, error) {
	esExecutionInfo := esResponse.Executions
	pinotExecutionInfo := pinotResponse.Executions

	vOfES := reflect.ValueOf(esExecutionInfo)
	typeOfesExecutionInfo := vOfES.Type()
	vOfPinot := reflect.ValueOf(pinotExecutionInfo)

	for i := 0; i < vOfES.NumField(); i++ {
		esFieldName := typeOfesExecutionInfo.Field(i).Name
		esValue := vOfES.Field(i).Interface()

		// if the value in ES is nil, then we don't need to compare
		if esValue == nil {
			continue
		}

		pinotValue := vOfPinot.Field(i).Interface()
		if strings.ToLower(esFieldName) == "searchattributes" {
			err := compareSearchAttributes(esValue, pinotValue)
			if err != nil {
				return nil, err
			}
		}

		if esValue != pinotValue {
			return nil,
				fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = %s, Pinot value = %s", esFieldName, esValue, pinotValue))
		}
	}

	return esResponse, nil
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

	return compareListWorkflowExecutionsResponse(esResponse, pinotResponse)
}
