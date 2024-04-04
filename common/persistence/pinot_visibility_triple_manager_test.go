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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log"
)

func TestNewPinotVisibilityTripleManager(t *testing.T) {
	// put this outside because need to use it as an input of the table tests
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager    VisibilityManager
		mockESVisibilityManager    VisibilityManager
		mockPinotVisibilityManager VisibilityManager
	}{
		"Case1: nil case": {
			mockDBVisibilityManager:    nil,
			mockESVisibilityManager:    nil,
			mockPinotVisibilityManager: nil,
		},
		"Case2: success case": {
			mockDBVisibilityManager:    NewMockVisibilityManager(ctrl),
			mockESVisibilityManager:    NewMockVisibilityManager(ctrl),
			mockPinotVisibilityManager: NewMockVisibilityManager(ctrl),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				NewPinotVisibilityTripleManager(test.mockDBVisibilityManager, test.mockPinotVisibilityManager,
					test.mockESVisibilityManager, nil, nil,
					nil, nil, log.NewNoop())
			})
		})
	}
}

func TestPinotTripleManagerClose(t *testing.T) {
	// put this outside because need to use it as an input of the table tests
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager              VisibilityManager
		mockPinotVisibilityManager           VisibilityManager
		mockESVisibilityManager              VisibilityManager
		mockDBVisibilityManagerAccordance    func(mockDBVisibilityManager *MockVisibilityManager)
		mockPinotVisibilityManagerAccordance func(mockPinotVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance    func(mockESVisibilityManager *MockVisibilityManager)
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
		"Case1-3: success case with pinot visibility is not nil": {
			mockPinotVisibilityManager: NewMockVisibilityManager(ctrl),
			mockPinotVisibilityManagerAccordance: func(mockPinotVisibilityManager *MockVisibilityManager) {
				mockPinotVisibilityManager.EXPECT().Close().Return().Times(1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockPinotVisibilityManager != nil {
				test.mockPinotVisibilityManagerAccordance(test.mockPinotVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}

			visibilityManager := NewPinotVisibilityTripleManager(test.mockDBVisibilityManager, test.mockPinotVisibilityManager,
				test.mockESVisibilityManager, nil, nil,
				nil, nil, log.NewNoop())
			assert.NotPanics(t, func() {
				visibilityManager.Close()
			})
		})
	}
}

func TestPinotTripleManagerGetName(t *testing.T) {
	// put this outside because need to use it as an input of the table tests
	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		mockDBVisibilityManager              VisibilityManager
		mockESVisibilityManager              VisibilityManager
		mockPinotVisibilityManager           VisibilityManager
		mockDBVisibilityManagerAccordance    func(mockDBVisibilityManager *MockVisibilityManager)
		mockPinotVisibilityManagerAccordance func(mockPinotVisibilityManager *MockVisibilityManager)
		mockESVisibilityManagerAccordance    func(mockESVisibilityManager *MockVisibilityManager)
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
		"Case1-3: success case with pinot visibility is not nil": {
			mockPinotVisibilityManager: NewMockVisibilityManager(ctrl),
			mockPinotVisibilityManagerAccordance: func(mockPinotVisibilityManager *MockVisibilityManager) {
				mockPinotVisibilityManager.EXPECT().GetName().Return(testTableName).Times(1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockDBVisibilityManager != nil {
				test.mockDBVisibilityManagerAccordance(test.mockDBVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockPinotVisibilityManager != nil {
				test.mockPinotVisibilityManagerAccordance(test.mockPinotVisibilityManager.(*MockVisibilityManager))
			}
			if test.mockESVisibilityManager != nil {
				test.mockESVisibilityManagerAccordance(test.mockESVisibilityManager.(*MockVisibilityManager))
			}
			visibilityManager := NewPinotVisibilityTripleManager(test.mockDBVisibilityManager, test.mockPinotVisibilityManager,
				test.mockESVisibilityManager, nil, nil,
				nil, nil, log.NewNoop())

			assert.NotPanics(t, func() {
				visibilityManager.GetName()
			})
		})
	}
}

func TestFilterAttrPrefix(t *testing.T) {
	tests := map[string]struct {
		expectedInput  string
		expectedOutput string
	}{
		"Case1: empty input": {
			expectedInput:  "",
			expectedOutput: "",
		},
		"Case2: filtered input": {
			expectedInput:  "`Attr.CustomIntField` = 12",
			expectedOutput: "CustomIntField = 12",
		},
		"Case3: complex input": {
			expectedInput:  "WorkflowID = 'test-wf' and (`Attr.CustomIntField` = 12 or `Attr.CustomStringField` = 'a-b-c' and WorkflowType = 'wf-type')",
			expectedOutput: "WorkflowID = 'test-wf' and (CustomIntField = 12 or CustomStringField = 'a-b-c' and WorkflowType = 'wf-type')",
		},
		"Case4: false positive case": {
			expectedInput:  "`Attr.CustomStringField` = '`Attr.ABCtesting'",
			expectedOutput: "CustomStringField = 'ABCtesting'", // this is supposed to be CustomStringField = '`Attr.ABCtesting'
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualOutput := filterAttrPrefix(test.expectedInput)
				assert.Equal(t, test.expectedOutput, actualOutput)
			})
		})
	}
}
