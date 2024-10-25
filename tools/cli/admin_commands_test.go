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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/tools/cli/clitest"
)

const (
	testShardID   = 1234
	testCluster   = "test-cluster"
	testQueueType = 2 // transfer queue
)

type cliTestData struct {
	serverFrontendClient *frontend.MockClient
	serverAdminClient    *admin.MockClient
	app                  *cli.App
}

func newCLITestData(t *testing.T) *cliTestData {
	var td cliTestData

	ctrl := gomock.NewController(t)

	td.serverFrontendClient = frontend.NewMockClient(ctrl)
	td.serverAdminClient = admin.NewMockClient(ctrl)

	// Set up the CLI app and mock dependencies
	td.app = NewCliApp(&clientFactoryMock{
		serverFrontendClient: td.serverFrontendClient,
		serverAdminClient:    td.serverAdminClient,
	})

	return &td
}

func TestAdminResetQueue(t *testing.T) {
	tests := []struct {
		name        string
		testSetup   func(td *cliTestData) *cli.Context
		errContains string // empty if no error is expected
	}{
		{
			name: "no shardID argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing cluster argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					/* no cluster argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing queue type argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					/* no queue type argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "all arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.serverAdminClient.EXPECT().ResetQueue(gomock.Any(), &types.ResetQueueRequest{
					ShardID:     testShardID,
					ClusterName: testCluster,
					Type:        common.Int32Ptr(testQueueType),
				})

				return cliCtx
			},
			errContains: "",
		},
		{
			name: "ResetQueue returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.serverAdminClient.EXPECT().ResetQueue(gomock.Any(), &types.ResetQueueRequest{
					ShardID:     testShardID,
					ClusterName: testCluster,
					Type:        common.Int32Ptr(testQueueType),
				}).Return(errors.New("critical error"))

				return cliCtx
			},
			errContains: "Failed to reset queue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)
			err := AdminResetQueue(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
		})
	}
}
