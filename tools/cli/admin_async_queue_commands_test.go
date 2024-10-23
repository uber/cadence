package cli

import (
	"errors"
	"flag"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/client/admin"
    "github.com/uber/cadence/client/frontend"
)

func TestAdminGetAsyncWFConfig(t *testing.T) {
    tests := []struct {
        name            string
        setup           func(*admin.MockClient)
        flags           map[string]string
        expectedError   string
        expectedOutput  string
    }{
        {
            name: "Success",
            setup: func(mockClient *admin.MockClient) {
                mockClient.EXPECT().GetDomainAsyncWorkflowConfiguraton(gomock.Any(), gomock.Any()).Return(&types.GetDomainAsyncWorkflowConfiguratonResponse{
                    Configuration: &types.AsyncWorkflowConfiguration{
                        Enabled: true,
                        PredefinedQueueName: "test-queue",
                        QueueType: "test-queue-type",
                    },
                }, nil)
            },
            flags: map[string]string{
                FlagDomain: "test-domain",
                FlagJSON:   `{"BufferSize": 10}`,
            },
            expectedOutput: "Async workflow queue config for domain test-domain:\n{\n  \"BufferSize\": 10\n}\n",
        },
        {
            name: "Missing Domain Flag",
            setup: func(mockClient *admin.MockClient) {},
            flags: map[string]string{},
            expectedError: "Required flag not present:: Option domain is required",
        },
        {
            name: "Config Not Found",
            setup: func(mockClient *admin.MockClient) {
                mockClient.EXPECT().GetDomainAsyncWorkflowConfiguraton(gomock.Any(), gomock.Any()).Return(nil, nil)
            },
            flags: map[string]string{
                FlagDomain: "test-domain",
            },
            expectedOutput: "Async workflow queue config not found for domain test-domain\n",
        },
        {
            name: "Error from Admin Client",
            setup: func(mockClient *admin.MockClient) {
                mockClient.EXPECT().GetDomainAsyncWorkflowConfiguraton(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
            },
            flags: map[string]string{
                FlagDomain: "test-domain",
            },
            expectedError: "Failed to get async wf queue config",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			serverFrontendClient := frontend.NewMockClient(mockCtrl)
			serverAdminClient := admin.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: serverFrontendClient,
				serverAdminClient:    serverAdminClient,
			})
			set := flag.NewFlagSet("test", 0)
            for k, v := range tt.flags {
                set.String(k, v, "")
            }
			c := cli.NewContext(app, set, nil)
            tt.setup(serverAdminClient)
			err := AdminGetAsyncWFConfig(c)
			if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)

            }
        })
    }
}

func TestAdminUpdateAsyncWFConfig(t *testing.T) {
    tests := []struct {
        name          string
        setup         func(*admin.MockClient)
        flags         map[string]string 
        expectedError string
        expectedOutput string
    }{
        {
            name: "Success",
            setup: func(mockClient *admin.MockClient) {
                mockClient.EXPECT().UpdateDomainAsyncWorkflowConfiguraton(gomock.Any(), gomock.Any()).Return(nil, nil)
            },
            flags: map[string]string{
                FlagDomain: "test-domain",
                FlagJSON:   `{"BufferSize": 10}`,
            },
            expectedOutput: "Successfully updated async workflow queue config for domain test-domain",
        },
        {
            name: "Missing Domain Flag",
            setup: func(mockClient *admin.MockClient) {},
            flags: map[string]string{
                FlagJSON: `{"BufferSize": 10}`,
            },
            expectedError: "Required flag not present:: Option domain is required",
        },
        {
            name: "Invalid JSON",
            setup: func(mockClient *admin.MockClient) {},
            flags: map[string]string{
                FlagDomain: "test-domain",
                FlagJSON:   "invalid-json",
            },
            expectedError: "Failed to parse async workflow config",
        },
        {
            name: "Error from Admin Client",
            setup: func(mockClient *admin.MockClient) {
                mockClient.EXPECT().UpdateDomainAsyncWorkflowConfiguraton(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
            },
            flags: map[string]string{
                FlagDomain: "test-domain",
                FlagJSON:   `{"BufferSize": 10}`,
            },
            expectedError: "Failed to update async workflow queue config",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			serverFrontendClient := frontend.NewMockClient(mockCtrl)
			serverAdminClient := admin.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: serverFrontendClient,
				serverAdminClient:    serverAdminClient,
			})
			set := flag.NewFlagSet("test", 0)
            for k, v := range tt.flags {
                set.String(k, v, "")
            }
			c := cli.NewContext(app, set, nil)
            tt.setup(serverAdminClient)
			err := AdminUpdateAsyncWFConfig(c)
			if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)

            }
        })
    }
}