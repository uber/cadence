package cli

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/client"
	"github.com/urfave/cli/v2"
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
