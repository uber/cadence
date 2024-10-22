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

package cli

import (
	"flag"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common/types"
)

func TestAdminDescribeTaskList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	expectedResponse := &types.DescribeTaskListResponse{
		Pollers: []*types.PollerInfo{
			{
				Identity: "test-poller",
			},
		},
		TaskListStatus: &types.TaskListStatus{
			BacklogCountHint: 10,
		},
	}
	serverFrontendClient.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(expectedResponse, nil).Times(1)

	c := setTasklistMock(app)
	err := AdminDescribeTaskList(c)
	assert.NoError(t, err)
}

func TestAdminDescribeTaskList_DescribeTaskListFails(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	serverFrontendClient.EXPECT().
		DescribeTaskList(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("DescribeTaskList failed")).
		Times(1)

	c := setTasklistMock(app)
	err := AdminDescribeTaskList(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Operation DescribeTaskList failed.")
}

func TestAdminDescribeTaskList_NoTaskListStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	expectedResponse := &types.DescribeTaskListResponse{
		Pollers:        []*types.PollerInfo{},
		TaskListStatus: nil,
	}

	serverFrontendClient.EXPECT().
		DescribeTaskList(gomock.Any(), gomock.Any()).
		Return(expectedResponse, nil).
		Times(1)

	c := setTasklistMock(app)
	err := AdminDescribeTaskList(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No tasklist status information.")
}

func TestAdminDescribeTaskList_NoPollers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	expectedResponse := &types.DescribeTaskListResponse{
		Pollers: []*types.PollerInfo{},
		TaskListStatus: &types.TaskListStatus{
			BacklogCountHint: 0,
		},
	}

	serverFrontendClient.EXPECT().
		DescribeTaskList(gomock.Any(), gomock.Any()).
		Return(expectedResponse, nil).
		Times(1)

	c := setTasklistMock(app)
	err := AdminDescribeTaskList(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No poller for tasklist: test-tasklist")
}

func TestAdminDescribeTaskList_GetRequiredOptionDomainError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	// Omit the Domain flag
	set := flag.NewFlagSet("test", 0)
	set.String(FlagTaskList, "test-tasklist", "TaskList flag")
	set.String(FlagTaskListType, "decision", "TaskListType flag")
	c := cli.NewContext(app, set, nil)

	err := AdminDescribeTaskList(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Required flag not found: ")
}

func TestAdminDescribeTaskList_GetRequiredOptionTaskListError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	// Omit the TaskList flag
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "Domain flag")
	set.String(FlagTaskListType, "decision", "TaskListType flag")
	c := cli.NewContext(app, set, nil)

	err := AdminDescribeTaskList(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Required flag not found: ")
}

func TestAdminDescribeTaskList_InvalidTaskListType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	// Set an invalid TaskListType
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "Domain flag")
	set.String(FlagTaskList, "test-tasklist", "TaskList flag")
	set.String(FlagTaskListType, "activity", "TaskListType flag")
	c := cli.NewContext(app, set, nil)

	expectedResponse := &types.DescribeTaskListResponse{
		Pollers: []*types.PollerInfo{
			{
				Identity: "test-poller",
			},
		},
		TaskListStatus: &types.TaskListStatus{
			BacklogCountHint: 10,
		},
	}
	serverFrontendClient.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(expectedResponse, nil).Times(1)

	err := AdminDescribeTaskList(c)
	assert.NoError(t, err)
}

func TestAdminListTaskList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Define table of test cases
	tests := []struct {
		name          string
		setupMocks    func(*frontend.MockClient)
		expectedError string
		domainFlag    string
		taskListFlag  string
		taskListType  string
	}{
		{
			name: "Success",
			setupMocks: func(client *frontend.MockClient) {
				expectedResponse := &types.GetTaskListsByDomainResponse{
					DecisionTaskListMap: map[string]*types.DescribeTaskListResponse{
						"decision-tasklist-1": {
							Pollers: []*types.PollerInfo{
								{Identity: "poller1"},
								{Identity: "poller2"},
							},
						},
						"decision-tasklist-2": {
							Pollers: []*types.PollerInfo{
								{Identity: "poller3"},
							},
						},
					},
					ActivityTaskListMap: map[string]*types.DescribeTaskListResponse{
						"activity-tasklist-1": {
							Pollers: []*types.PollerInfo{
								{Identity: "poller4"},
							},
						},
					},
				}
				client.EXPECT().
					GetTaskListsByDomain(gomock.Any(), gomock.Any()).
					Return(expectedResponse, nil).
					Times(1)
			},
			expectedError: "",
			domainFlag:    "test-domain",
			taskListFlag:  "test-tasklist",
			taskListType:  "decision",
		},
		{
			name: "GetTaskListsByDomainFails",
			setupMocks: func(client *frontend.MockClient) {
				client.EXPECT().
					GetTaskListsByDomain(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("GetTaskListsByDomain failed")).
					Times(1)
			},
			expectedError: "Operation GetTaskListByDomain failed",
			domainFlag:    "test-domain",
			taskListFlag:  "test-tasklist",
			taskListType:  "decision",
		},
		{
			name:          "NoDomainFlag",
			setupMocks:    func(client *frontend.MockClient) {},
			expectedError: "Required flag not found",
			domainFlag:    "", // Omit Domain flag
			taskListFlag:  "test-tasklist",
			taskListType:  "decision",
		},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock clients
			serverFrontendClient := frontend.NewMockClient(mockCtrl)
			serverAdminClient := admin.NewMockClient(mockCtrl)

			// Create CLI app with mock clients
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: serverFrontendClient,
				serverAdminClient:    serverAdminClient,
			})

			// Set up mocks for the current test case
			tt.setupMocks(serverFrontendClient)

			// Set up the CLI context
			set := flag.NewFlagSet("test", 0)
			if tt.domainFlag != "" {
				set.String(FlagDomain, tt.domainFlag, "Domain flag")
			}
			set.String(FlagTaskList, tt.taskListFlag, "TaskList flag")
			set.String(FlagTaskListType, tt.taskListType, "TaskListType flag")
			c := cli.NewContext(app, set, nil)

			// Call the function under test
			err := AdminListTaskList(c)

			// Check expected outcomes
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to set up the CLI context
func setTasklistMock(app *cli.App) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "Domain flag")
	set.String(FlagTaskList, "test-tasklist", "TaskList flag")
	set.String(FlagTaskListType, "decision", "TaskListType flag")
	c := cli.NewContext(app, set, nil)
	return c
}
