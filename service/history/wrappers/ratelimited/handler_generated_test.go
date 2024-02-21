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

package ratelimited

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/handler"
	"github.com/uber/cadence/service/history/workflowcache"
)

const (
	testDomainID   = "test-domain-id"
	testWorkflowID = "test-workflow-id"
)

func TestRatelimitedEndpoints_Table(t *testing.T) {
	controller := gomock.NewController(t)

	workflowIDCache := workflowcache.NewMockWFCache(controller)
	handlerMock := handler.NewMockHandler(controller)

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
			// For now true and false needs to do the same as we are only shadowing
			workflowIDCache.EXPECT().AllowExternal(testWorkflowID, testDomainID).Return(true).Times(1)
			tt.mock()
			_, err := tt.call()
			assert.NoError(t, err)

			workflowIDCache.EXPECT().AllowExternal(testWorkflowID, testDomainID).Return(false).Times(1)
			tt.mock()
			_, err = tt.call()
			assert.NoError(t, err)
		})
	}
}
