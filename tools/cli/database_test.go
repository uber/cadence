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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/client"
)

func TestDefaultManagerFactory(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*client.MockFactory, *gomock.Controller) interface{}
		methodToTest func(*defaultManagerFactory, *cli.Context) (interface{}, error)
		expectError  bool
	}{
		{
			name: "initializeExecutionManager - success",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
				mockFactory.EXPECT().NewExecutionManager(gomock.Any()).Return(mockExecutionManager, nil)
				return mockExecutionManager
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeExecutionManager(ctx, 1)
			},
			expectError: false,
		},
		{
			name: "initializeExecutionManager - error",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockFactory.EXPECT().NewExecutionManager(gomock.Any()).Return(nil, fmt.Errorf("some error"))
				return nil
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeExecutionManager(ctx, 1)
			},
			expectError: true,
		},
		{
			name: "initializeHistoryManager - success",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
				mockFactory.EXPECT().NewHistoryManager().Return(mockHistoryManager, nil)
				return mockHistoryManager
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeHistoryManager(ctx)
			},
			expectError: false,
		},
		{
			name: "initializeHistoryManager - error",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockFactory.EXPECT().NewHistoryManager().Return(nil, fmt.Errorf("some error"))
				return nil
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeHistoryManager(ctx)
			},
			expectError: true,
		},
		{
			name: "initializeShardManager - success",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockShardManager := persistence.NewMockShardManager(ctrl)
				mockFactory.EXPECT().NewShardManager().Return(mockShardManager, nil)
				return mockShardManager
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeShardManager(ctx)
			},
			expectError: false,
		},
		{
			name: "initializeShardManager - error",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockFactory.EXPECT().NewShardManager().Return(nil, fmt.Errorf("some error"))
				return nil
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeShardManager(ctx)
			},
			expectError: true,
		},
		{
			name: "initializeDomainManager - success",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockDomainManager := persistence.NewMockDomainManager(ctrl)
				mockFactory.EXPECT().NewDomainManager().Return(mockDomainManager, nil)
				return mockDomainManager
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeDomainManager(ctx)
			},
			expectError: false,
		},
		{
			name: "initializeDomainManager - error",
			setupMocks: func(mockFactory *client.MockFactory, ctrl *gomock.Controller) interface{} {
				mockFactory.EXPECT().NewDomainManager().Return(nil, fmt.Errorf("some error"))
				return nil
			},
			methodToTest: func(f *defaultManagerFactory, ctx *cli.Context) (interface{}, error) {
				return f.initializeDomainManager(ctx)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFactory := client.NewMockFactory(ctrl)
			expectedManager := tt.setupMocks(mockFactory, ctrl)

			f := &defaultManagerFactory{persistenceFactory: mockFactory}
			app := &cli.App{}
			ctx := cli.NewContext(app, nil, nil)

			result, err := tt.methodToTest(f, ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedManager, result)
			}
		})
	}
}
