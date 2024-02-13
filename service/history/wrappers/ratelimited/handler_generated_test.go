package ratelimited

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history"
	"github.com/uber/cadence/service/history/workflowcache"
)

const (
	testDomainID   = "test-domain-id"
	testWorkflowID = "test-workflow-id"
)

func TestRatelimitedEndpoints_Table(t *testing.T) {
	controller := gomock.NewController(t)

	workflowIDCache := workflowcache.NewMockWFCache(controller)
	handlerMock := history.NewMockHandler(controller)

	wrapper := NewHistoryHandler(handlerMock, workflowIDCache)

	tests := []struct {
		name string
		call func() (interface{}, error)
		mock func()
	}{
		{
			name: "StartWorkflowExecution",
			call: func() (interface{}, error) {
				startRequest := &types.HistoryStartWorkflowExecutionRequest{
					DomainUUID:   testDomainID,
					StartRequest: &types.StartWorkflowExecutionRequest{WorkflowID: testWorkflowID},
				}
				return wrapper.StartWorkflowExecution(context.Background(), startRequest)
			},
			mock: func() {
				handlerMock.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
		},
		{
			name: "SignalWithStartWorkflowExecution",
			call: func() (interface{}, error) {
				signalWithStartRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{
					DomainUUID:             testDomainID,
					SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{WorkflowID: testWorkflowID},
				}

				return wrapper.SignalWithStartWorkflowExecution(context.Background(), signalWithStartRequest)
			},
			mock: func() {
				handlerMock.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
		},
		{
			name: "SignalWorkflowExecution",
			call: func() (interface{}, error) {
				signalRequest := &types.HistorySignalWorkflowExecutionRequest{
					DomainUUID: testDomainID,
					SignalRequest: &types.SignalWorkflowExecutionRequest{
						WorkflowExecution: &types.WorkflowExecution{WorkflowID: testWorkflowID},
					},
				}

				return nil, wrapper.SignalWorkflowExecution(context.Background(), signalRequest)
			},
			mock: func() {
				handlerMock.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
		},
		{
			name: "DescribeWorkflowExecution",
			call: func() (interface{}, error) {
				describeRequest := &types.HistoryDescribeWorkflowExecutionRequest{
					DomainUUID: testDomainID,
					Request: &types.DescribeWorkflowExecutionRequest{
						Execution: &types.WorkflowExecution{WorkflowID: testWorkflowID},
					},
				}

				return wrapper.DescribeWorkflowExecution(context.Background(), describeRequest)
			},
			mock: func() {
				handlerMock.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowIDCache.EXPECT().AllowExternal(testWorkflowID, testDomainID).Return(true).Times(1)
			tt.mock()
			_, err := tt.call()
			assert.NoError(t, err)

			workflowIDCache.EXPECT().AllowExternal(testWorkflowID, testDomainID).Return(false).Times(1)
			_, err = tt.call()
			assert.Error(t, err)
		})
	}
}
