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

package persistence

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"testing"

	"github.com/uber/cadence/common/log"
)

func TestNewVisibilityDualManager(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager VisibilityManager
		mockESVisibilityManager VisibilityManager
	}{
		"Case1: nil case": {
			mockDBVisibilityManager: nil,
			mockESVisibilityManager: nil,
		},
		"Case2: success case": {
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				NewVisibilityDualManager(test.mockDBVisibilityManager, test.mockESVisibilityManager, nil, nil, log.NewNoop())
			})
		})
	}
}

func TestDualManagerClose(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
	}{
		"Case1-1: success case with DB visibility is not nil": {
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().Close().Return().Times(1)
			},
		},
		"Case1-2: success case with ES visibility is not nil": {
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().Close().Return().Times(1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager, test.mockESVisibilityManager, nil, nil, log.NewNoop())
			assert.NotPanics(t, func() {
				visibilityManager.Close()
			})
		})
	}
}

func TestDualManagerGetName(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
	}{
		"Case1-1: success case with DB visibility is not nil": {
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().GetName().Return(testTableName).Times(1)

			},
		},
		"Case1-2: success case with ES visibility is not nil": {
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().GetName().Return(testTableName).Times(1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager, test.mockESVisibilityManager, nil, nil, log.NewNoop())

			assert.NotPanics(t, func() {
				visibilityManager.GetName()
			})
		})
	}
}

func TestDualRecordWorkflowExecutionStarted(t *testing.T) {
	request := &RecordWorkflowExecutionStartedRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		request                           *RecordWorkflowExecutionStartedRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case1-1: success case with DB visibility is not nil": {
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionStarted(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionStarted(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.RecordWorkflowExecutionStarted(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDualRecordWorkflowExecutionClosed(t *testing.T) {
	request := &RecordWorkflowExecutionClosedRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		context                           context.Context
		request                           *RecordWorkflowExecutionClosedRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case0-1: error case with advancedVisibilityWritingMode is nil": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case0-2: error case with ES has errors in dual mode": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error")).AnyTimes()
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeDual),
			expectedError:                 fmt.Errorf("error"),
		},
		"Case1-1: success case with DB visibility is not nil": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			context:                 context.Background(),
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
		"Case1-3: success case with dual manager": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeDual),
			expectedError:                 nil,
		},
		"Case2-1: choose DB visibility manager when it is nil": {
			context:                 context.Background(),
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
		},
		"Case2-2: choose ES visibility manager when it is nil": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
		},
		"Case2-3: choose both when ES is nil": {
			context:                 context.Background(),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeDual),
		},
		"Case2-4: choose both when DB is nil": {
			context:                 context.Background(),
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeDual),
		},
		"Case3-1: chooseVisibilityModeForAdmin when ES is nil": {
			context:                 context.WithValue(context.Background(), VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
		},
		"Case3-2: chooseVisibilityModeForAdmin when DB is nil": {
			context:                 context.WithValue(context.Background(), VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
		},
		"Case3-3: chooseVisibilityModeForAdmin when both are not nil": {
			context:                 context.WithValue(context.Background(), VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.RecordWorkflowExecutionClosed(test.context, test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// test an edge case
func TestChooseVisibilityModeForAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	dbManager := NewMockVisibilityManager(ctrl)
	esManager := NewMockVisibilityManager(ctrl)
	mgr := NewVisibilityDualManager(dbManager, esManager, nil, nil, log.NewNoop())
	dualManager := mgr.(*visibilityDualManager)
	dualManager.dbVisibilityManager = nil
	dualManager.esVisibilityManager = nil
	assert.Equal(t, "INVALID_ADMIN_MODE", dualManager.chooseVisibilityModeForAdmin())
}

func TestDualRecordWorkflowExecutionUninitialized(t *testing.T) {
	request := &RecordWorkflowExecutionUninitializedRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		request                           *RecordWorkflowExecutionUninitializedRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case1-1: success case with DB visibility is not nil": {
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().RecordWorkflowExecutionUninitialized(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().RecordWorkflowExecutionUninitialized(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.RecordWorkflowExecutionUninitialized(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDualUpsertWorkflowExecution(t *testing.T) {
	request := &UpsertWorkflowExecutionRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		request                           *UpsertWorkflowExecutionRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case1-1: success case with DB visibility is not nil": {
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().UpsertWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().UpsertWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.UpsertWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDualDeleteWorkflowExecution(t *testing.T) {
	request := &VisibilityDeleteWorkflowExecutionRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		request                           *VisibilityDeleteWorkflowExecutionRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case1-1: success case with DB visibility is not nil": {
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.DeleteWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDualDeleteUninitializedWorkflowExecution(t *testing.T) {
	request := &VisibilityDeleteWorkflowExecutionRequest{}

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		request                           *VisibilityDeleteWorkflowExecutionRequest
		mockDBVisibilityManager           VisibilityManager
		mockESVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance func(mockDBVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance func(mockESVisibilityManager *MockVisibilityManager)
		advancedVisibilityWritingMode     dynamicconfig.StringPropertyFn
		expectedError                     error
	}{
		"Case1-1: success case with DB visibility is not nil": {
			request:                 request,
			mockDBVisibilityManager: NewMockVisibilityManager(ctrl),
			mockDBVisibilityManagerAccordance: func(mockDBVisibilityManager *MockVisibilityManager) {
				mockDBVisibilityManager.EXPECT().DeleteUninitializedWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOff),
			expectedError:                 nil,
		},
		"Case1-2: success case with ES visibility is not nil": {
			request:                 request,
			mockESVisibilityManager: NewMockVisibilityManager(ctrl),
			mockESVisibilityManagerAccordance: func(mockESVisibilityManager *MockVisibilityManager) {
				mockESVisibilityManager.EXPECT().DeleteUninitializedWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			advancedVisibilityWritingMode: dynamicconfig.GetStringPropertyFn(common.AdvancedVisibilityWritingModeOn),
			expectedError:                 nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewVisibilityDualManager(test.mockDBVisibilityManager,
				test.mockESVisibilityManager, nil, test.advancedVisibilityWritingMode, log.NewNoop())

			err := visibilityManager.DeleteUninitializedWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

/*




func TestListOpenWorkflowExecutions(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsRequest{}

	request := &ListWorkflowExecutionsRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListOpenWorkflowExecutions(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListClosedWorkflowExecutions(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsRequest{}

	request := &ListWorkflowExecutionsRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListClosedWorkflowExecutions(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListOpenWorkflowExecutionsByType(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByTypeRequest{}

	request := &ListWorkflowExecutionsByTypeRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByTypeRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListOpenWorkflowExecutionsByType(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTestListClosedWorkflowExecutionsByType(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByTypeRequest{}

	request := &ListWorkflowExecutionsByTypeRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByTypeRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListClosedWorkflowExecutionsByType(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListOpenWorkflowExecutionsByWorkflowID(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByWorkflowIDRequest{}

	request := &ListWorkflowExecutionsByWorkflowIDRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByWorkflowIDRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListOpenWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListOpenWorkflowExecutionsByWorkflowID(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListClosedWorkflowExecutionsByWorkflowID(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByWorkflowIDRequest{}

	request := &ListWorkflowExecutionsByWorkflowIDRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByWorkflowIDRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListClosedWorkflowExecutionsByWorkflowID(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListClosedWorkflowExecutionsByStatus(t *testing.T) {
	errorRequest := &ListClosedWorkflowExecutionsByStatusRequest{}

	request := &ListClosedWorkflowExecutionsByStatusRequest{}

	tests := map[string]struct {
		request                   *ListClosedWorkflowExecutionsByStatusRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByStatus(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListClosedWorkflowExecutionsByStatus(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListClosedWorkflowExecutionsByStatus(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetClosedWorkflowExecution(t *testing.T) {
	errorRequest := &GetClosedWorkflowExecutionRequest{}

	request := &GetClosedWorkflowExecutionRequest{}

	tests := map[string]struct {
		request                   *GetClosedWorkflowExecutionRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().GetClosedWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().GetClosedWorkflowExecution(gomock.Any(), gomock.Any()).Return(
					&InternalGetClosedWorkflowExecutionResponse{
						Execution: &InternalVisibilityWorkflowExecutionInfo{
							Memo: &DataBlob{
								Data: []byte("test string"),
							},
						},
					}, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.GetClosedWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListWorkflowExecutions(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByQueryRequest{}

	request := &ListWorkflowExecutionsByQueryRequest{}

	testStatus := types.WorkflowExecutionCloseStatus(2)

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByQueryRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(
					&InternalListWorkflowExecutionsResponse{
						Executions: []*InternalVisibilityWorkflowExecutionInfo{
							{
								Status: &testStatus,
								Memo: &DataBlob{
									Data: []byte("test string"),
								},
								SearchAttributes: map[string]interface{}{
									"CustomStringField": []byte("test string"),
									"CustomTimeField":   []byte("2020-01-01T00:00:00Z"),
								},
							},
						}}, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ListWorkflowExecutions(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScanWorkflowExecutions(t *testing.T) {
	errorRequest := &ListWorkflowExecutionsByQueryRequest{}

	request := &ListWorkflowExecutionsByQueryRequest{}

	tests := map[string]struct {
		request                   *ListWorkflowExecutionsByQueryRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.ScanWorkflowExecutions(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCountWorkflowExecutions(t *testing.T) {
	errorRequest := &CountWorkflowExecutionsRequest{}

	request := &CountWorkflowExecutionsRequest{}

	tests := map[string]struct {
		request                   *CountWorkflowExecutionsRequest
		visibilityStoreAffordance func(mockVisibilityStore *MockVisibilityStore)
		expectedError             error
	}{
		"Case1: error case": {
			request: errorRequest,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request: request,
			visibilityStoreAffordance: func(mockVisibilityStore *MockVisibilityStore) {
				mockVisibilityStore.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			test.visibilityStoreAffordance(mockVisibilityStore)

			_, err := visibilityManager.CountWorkflowExecutions(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSearchAttributes(t *testing.T) {
	tests := map[string]struct {
		input          *map[string]interface{}
		expectedOutput *types.SearchAttributes
		expectedError  error
	}{
		"Case1: error case": {
			input: &map[string]interface{}{
				"CustomStringField": make(chan int),
			},
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			input:         &map[string]interface{}{},
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			visibilityManagerImpl := visibilityManager.(*visibilityManagerImpl)

			actualOutput, actualErr := visibilityManagerImpl.getSearchAttributes(*test.input)
			if test.expectedError != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
				assert.NotNil(t, actualOutput)
			}
		})
	}
}

func TestConvertVisibilityWorkflowExecutionInfo(t *testing.T) {
	testExecutionTime := int64(0)
	testStartTime := time.Now()
	testResultTime := testStartTime.UnixNano()

	tests := map[string]struct {
		input          *InternalVisibilityWorkflowExecutionInfo
		expectedOutput *types.WorkflowExecutionInfo
	}{
		"Case1: error case": {
			input: &InternalVisibilityWorkflowExecutionInfo{
				SearchAttributes: map[string]interface{}{
					"CustomStringField": make(chan int),
				},
				ExecutionTime: time.UnixMilli(int64(testExecutionTime)),
				StartTime:     testStartTime,
			},
			expectedOutput: &types.WorkflowExecutionInfo{
				ExecutionTime: &testResultTime,
			},
		},
		"Case2: normal case with Execution Time is 0": {
			input: &InternalVisibilityWorkflowExecutionInfo{
				ExecutionTime: time.UnixMilli(int64(testExecutionTime)),
				StartTime:     testStartTime,
			},
			expectedOutput: &types.WorkflowExecutionInfo{
				ExecutionTime: &testResultTime,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			visibilityManagerImpl := visibilityManager.(*visibilityManagerImpl)

			actualOutput := visibilityManagerImpl.convertVisibilityWorkflowExecutionInfo(test.input)
			assert.Equal(t, *test.expectedOutput.ExecutionTime, *actualOutput.ExecutionTime)
		})
	}
}

func TestToInternalListWorkflowExecutionsRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVisibilityStore := NewMockVisibilityStore(ctrl)
	visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
	visibilityManagerImpl := visibilityManager.(*visibilityManagerImpl)

	assert.Nil(t, visibilityManagerImpl.toInternalListWorkflowExecutionsRequest(nil))
}

func TestSerializeMemo(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVisibilityStore := NewMockVisibilityStore(ctrl)
	mockPayloadSerializer := NewMockPayloadSerializer(ctrl)
	mockPayloadSerializer.EXPECT().SerializeVisibilityMemo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)
	visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
	visibilityManagerImpl := visibilityManager.(*visibilityManagerImpl)
	visibilityManagerImpl.serializer = mockPayloadSerializer
	assert.NotPanics(t, func() {
		visibilityManagerImpl.serializeMemo(nil, "testDomain", "testWorkflowID", "testRunID")
	})
}
*/
