// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
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
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/visibility"
	"github.com/uber/cadence/service/worker/failovermanager"
)

func TestAdminAddSearchAttribute_isValueTypeValid(t *testing.T) {
	testCases := []struct {
		name     string
		input    int
		expected bool
	}{
		{
			name:     "negative",
			input:    -1,
			expected: false,
		},
		{
			name:     "valid",
			input:    0,
			expected: true,
		},
		{
			name:     "valid",
			input:    5,
			expected: true,
		},
		{
			name:     "unknown",
			input:    6,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, isValueTypeValid(testCase.input))
	}
}

func TestAdminFailover(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	domainCLI := &domainCLIImpl{
		frontendClient: serverFrontendClient,
	}

	var listDomainsResponse = &types.ListDomainsResponse{
		Domains: []*types.DescribeDomainResponse{
			{
				DomainInfo: &types.DomainInfo{
					Name:        "test-domain",
					Description: "a test domain",
					OwnerEmail:  "test@uber.com",
					Data: map[string]string{
						common.DomainDataKeyForManagedFailover: "true",
					},
				},
				ReplicationConfiguration: &types.DomainReplicationConfiguration{
					ActiveClusterName: "active",
					Clusters: []*types.ClusterReplicationConfiguration{
						{
							ClusterName: "active",
						},
						{
							ClusterName: "standby",
						},
					},
				},
			},
		},
	}

	serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResponse, nil).Times(1)
	serverFrontendClient.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	set := flag.NewFlagSet("test", 0)
	set.String(FlagActiveClusterName, "standby", "test flag")

	cliContext := cli.NewContext(nil, set, nil)
	succeed, failed, err := domainCLI.failoverDomains(cliContext)
	assert.Equal(t, []string{"test-domain"}, succeed)
	assert.Equal(t, 0, len(failed))
	assert.NoError(t, err)

	serverFrontendClient.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(listDomainsResponse, nil).Times(1)
	set = flag.NewFlagSet("test", 0)
	set.String(FlagActiveClusterName, "active", "test flag")

	cliContext = cli.NewContext(nil, set, nil)
	succeed, failed, err = domainCLI.failoverDomains(cliContext)
	assert.Equal(t, 0, len(succeed))
	assert.Equal(t, 0, len(failed))
	assert.NoError(t, err)
}

func TestValidSearchAttributeKey(t *testing.T) {
	assert.NoError(t, visibility.ValidateSearchAttributeKey("city"))
	assert.NoError(t, visibility.ValidateSearchAttributeKey("cityId"))
	assert.NoError(t, visibility.ValidateSearchAttributeKey("paymentProfileUUID"))
	assert.NoError(t, visibility.ValidateSearchAttributeKey("job_type"))

	assert.Error(t, visibility.ValidateSearchAttributeKey("payments-biling-invoices-TransactionUUID"))
	assert.Error(t, visibility.ValidateSearchAttributeKey("9lives"))
	assert.Error(t, visibility.ValidateSearchAttributeKey("tax%"))
}

func TestAdminDescribeCluster(t *testing.T) {
	// Initialize mock controller
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Create mock frontend and admin clients
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)

	// Set up the CLI app and mock dependencies
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func() {
				// Expected response from DescribeCluster
				expectedResponse := &types.DescribeClusterResponse{
					SupportedClientVersions: &types.SupportedClientVersions{
						GoSdk: "1.5.0",
					},
				}
				// Mock the DescribeCluster call
				serverAdminClient.EXPECT().DescribeCluster(gomock.Any()).Return(expectedResponse, nil).Times(1)
			},
			expectedError: "",
		},
		{
			name: "DescribeClusterError",
			mockSetup: func() {
				// Mock DescribeCluster to return an error
				serverAdminClient.EXPECT().DescribeCluster(gomock.Any()).Return(nil, fmt.Errorf("DescribeCluster failed")).Times(1)
			},
			expectedError: "Operation DescribeCluster failed.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock based on the specific test case
			tt.mockSetup()

			// Set up CLI context
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// Call AdminDescribeCluster
			err := AdminDescribeCluster(c)

			// Check the result based on the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdminRebalanceStart(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(mockFrontClient *frontend.MockClient, mockClientFactory *MockClientFactory)
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func(mockFrontClient *frontend.MockClient, mockClientFactory *MockClientFactory) {
				// Mock StartWorkflowExecution response
				mockResponse := &types.StartWorkflowExecutionResponse{
					RunID: "test-run-id",
				}
				mockClientFactory.EXPECT().ServerFrontendClient(gomock.Any()).Return(mockFrontClient, nil).Times(1)
				mockFrontClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(mockResponse, nil).Times(1)
			},
			expectedError: "",
		},
		{
			name: "ServerFrontendClientError",
			mockSetup: func(mockFrontClient *frontend.MockClient, mockClientFactory *MockClientFactory) {
				// Mock ServerFrontendClient to return an error
				mockClientFactory.EXPECT().ServerFrontendClient(gomock.Any()).Return(nil, fmt.Errorf("failed to get frontend client")).Times(1)
			},
			expectedError: "failed to get frontend client",
		},
		{
			name: "StartWorkflowExecutionError",
			mockSetup: func(mockFrontClient *frontend.MockClient, mockClientFactory *MockClientFactory) {
				// Mock StartWorkflowExecution to return an error
				mockClientFactory.EXPECT().ServerFrontendClient(gomock.Any()).Return(mockFrontClient, nil).Times(1)
				mockFrontClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to start workflow")).Times(1)
			},
			expectedError: "Failed to start failover workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock controller
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// Create mock Cadence client and client factory
			mockFrontClient := frontend.NewMockClient(mockCtrl)
			mockClientFactory := NewMockClientFactory(mockCtrl)

			// Set up the CLI app
			app := cli.NewApp()
			app.Metadata = map[string]interface{}{
				"deps": &deps{
					ClientFactory: mockClientFactory,
				},
			}

			// Set up the mocks for the specific test case
			tt.mockSetup(mockFrontClient, mockClientFactory)

			// Use setContextMock to set the CLI context
			c := setContextMock(app)

			// Call AdminRebalanceStart
			err := AdminRebalanceStart(c)

			// Check the result based on the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to set up the CLI context for AdminRebalanceStart
func setContextMock(app *cli.App) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	// You can add any flags relevant to AdminRebalanceStart here
	set.String(FlagDomain, "test-domain", "Domain flag")

	c := cli.NewContext(app, set, nil)
	return c
}

func TestIntValTypeToString(t *testing.T) {
	tests := []struct {
		name     string
		valType  int
		expected string
	}{
		{
			name:     "StringType",
			valType:  0,
			expected: "String",
		},
		{
			name:     "KeywordType",
			valType:  1,
			expected: "Keyword",
		},
		{
			name:     "IntType",
			valType:  2,
			expected: "Int",
		},
		{
			name:     "DoubleType",
			valType:  3,
			expected: "Double",
		},
		{
			name:     "BoolType",
			valType:  4,
			expected: "Bool",
		},
		{
			name:     "DatetimeType",
			valType:  5,
			expected: "Datetime",
		},
		{
			name:     "UnknownType",
			valType:  999,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function and check the result
			result := intValTypeToString(tt.valType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminRebalanceList(t *testing.T) {
	tests := []struct {
		name          string
		prepareEnv    func() *cli.Context
		expectedError string
	}{
		{
			name: "Success",
			prepareEnv: func() *cli.Context {
				// Initialize the mock client factory and frontend client
				mockFrontClient := frontend.NewMockClient(gomock.NewController(t))
				mockClientFactory := NewMockClientFactory(gomock.NewController(t))

				// Mock successful ListWorkflow call
				mockClientFactory.EXPECT().ServerFrontendClient(gomock.Any()).Return(mockFrontClient, nil).Times(1)
				mockFrontClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{}, nil).Times(1)
				mockFrontClient.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.ListClosedWorkflowExecutionsResponse{}, nil).Times(1)

				// Create CLI app and set up flag set
				app := cli.NewApp()
				app.Metadata = map[string]interface{}{
					"deps": &deps{
						ClientFactory: mockClientFactory,
					},
				}
				set := flag.NewFlagSet("test", 0)
				set.String(FlagWorkflowID, "", "workflow ID flag")
				set.String(FlagDomain, "", "domain flag")
				c := cli.NewContext(app, set, nil)

				// Set flags for workflow ID and domain
				_ = c.Set(FlagWorkflowID, failovermanager.RebalanceWorkflowID)
				_ = c.Set(FlagDomain, common.SystemLocalDomainName)

				return c
			},
			expectedError: "",
		},
		{
			name: "SetWorkflowIDError",
			prepareEnv: func() *cli.Context {
				// Create CLI app and set up flag set without FlagWorkflowID
				app := cli.NewApp()
				set := flag.NewFlagSet("test", 0)
				set.String(FlagDomain, "", "domain flag") // Only Domain flag is set
				c := cli.NewContext(app, set, nil)

				// Set only the domain flag, so setting FlagWorkflowID should trigger an error
				_ = c.Set(FlagDomain, common.SystemLocalDomainName)

				return c
			},
			expectedError: "no such flag -workflow_id",
		},
		{
			name: "SetDomainError",
			prepareEnv: func() *cli.Context {
				// Create CLI app and set up flag set without FlagDomain
				app := cli.NewApp()
				set := flag.NewFlagSet("test", 0)
				set.String(FlagWorkflowID, "", "workflow ID flag") // Only Workflow ID flag is set
				c := cli.NewContext(app, set, nil)

				// Set workflow ID flag, but not the domain flag
				_ = c.Set(FlagWorkflowID, failovermanager.RebalanceWorkflowID)

				return c
			},
			expectedError: "no such flag -domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the test environment for the specific test case
			c := tt.prepareEnv()

			// Call AdminRebalanceList
			err := AdminRebalanceList(c)

			// Check the result based on the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdminAddSearchAttribute_errors(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(app *cli.App) *cli.Context
		expectedError string
	}{
		{
			name: "MissingSearchAttributesKey",
			setupContext: func(app *cli.App) *cli.Context {
				// Simulate missing FlagSearchAttributesKey
				set := flag.NewFlagSet("test", 0)
				// No FlagSearchAttributesKey set
				return cli.NewContext(app, set, nil)
			},
			expectedError: "Required flag not present:",
		},
		{
			name: "InvalidSearchAttributeKey",
			setupContext: func(app *cli.App) *cli.Context {
				// Provide an invalid key to trigger ValidateSearchAttributeKey error
				set := flag.NewFlagSet("test", 0)
				set.String(FlagSearchAttributesKey, "123_invalid_key", "Key flag") // Invalid key, starts with number
				return cli.NewContext(app, set, nil)
			},
			expectedError: "Invalid search-attribute key.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI app
			app := cli.NewApp()

			// Set up the CLI context for the specific test case
			c := tt.setupContext(app)

			// Call AdminAddSearchAttribute
			err := AdminAddSearchAttribute(c)

			// Check the result based on the expected outcome
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
