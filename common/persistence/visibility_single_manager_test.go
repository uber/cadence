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
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/uber/cadence/common/log"
)

func TestNewVisibilityManagerImpl(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			assert.NotPanics(t, func() {
				NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			})
		})
	}
}

func TestClose(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			mockVisibilityStore.EXPECT().Close().Return().Times(1)

			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			assert.NotPanics(t, func() {
				visibilityManager.Close()
			})
		})
	}
}

func TestGetName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			mockVisibilityStore.EXPECT().GetName().Return(testTableName).Times(1)

			NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			assert.NotPanics(t, func() {
				visibilityManager.GetName()
			})
		})
	}
}

//func TestRecordWorkflowExecutionStarted(t *testing.T) {
//	// test non-empty request fields match
//	errorRequest := &p.InternalRecordWorkflowExecutionStartedRequest{
//		WorkflowID: "wid",
//		Memo:       p.NewDataBlob([]byte(`test bytes`), common.EncodingTypeThriftRW),
//		SearchAttributes: map[string][]byte{
//			"CustomStringField": []byte("test string"),
//			"CustomTimeField":   []byte("2020-01-01T00:00:00Z"),
//		},
//	}
//
//	request := &p.InternalRecordWorkflowExecutionStartedRequest{
//		WorkflowID: "wid",
//		Memo:       p.NewDataBlob([]byte(`test bytes`), common.EncodingTypeThriftRW),
//	}
//
//	customStringField, err := json.Marshal("test string")
//	assert.NoError(t, err)
//	requestWithSearchAttributes := &p.InternalRecordWorkflowExecutionStartedRequest{
//		WorkflowID: "wid",
//		Memo:       p.NewDataBlob([]byte(`test bytes`), common.EncodingTypeThriftRW),
//		SearchAttributes: map[string][]byte{
//			"CustomStringField": customStringField,
//		},
//	}
//
//	tests := map[string]struct {
//		request                *p.InternalRecordWorkflowExecutionStartedRequest
//		producerMockAffordance func(mockProducer *mocks.KafkaProducer)
//		expectedError          error
//	}{
//		"Case1: error case": {
//			request: errorRequest,
//			producerMockAffordance: func(mockProducer *mocks.KafkaProducer) {
//				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
//					return true
//				})).Return(&types.BadRequestError{}).Once()
//			},
//			expectedError: fmt.Errorf("invalid character 'e' in literal true (expecting 'r')"),
//		},
//		"Case2: normal case": {
//			request: request,
//			producerMockAffordance: func(mockProducer *mocks.KafkaProducer) {
//				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
//					assert.Equal(t, request.WorkflowID, input.GetWorkflowID())
//					return true
//				})).Return(nil).Once()
//			},
//			expectedError: nil,
//		},
//		"Case3: normal case with search attributes": {
//			request: requestWithSearchAttributes,
//			producerMockAffordance: func(mockProducer *mocks.KafkaProducer) {
//				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
//					assert.Equal(t, request.WorkflowID, input.GetWorkflowID())
//					return true
//				})).Return(nil).Once()
//			},
//			expectedError: nil,
//		},
//	}
//
//	for name, test := range tests {
//		t.Run(name, func(t *testing.T) {
//			ctrl := gomock.NewController(t)
//			mockPinotClient := pnt.NewMockGenericClient(ctrl)
//			mockProducer := &mocks.KafkaProducer{}
//			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
//				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
//				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
//			}, mockProducer, log.NewNoop())
//			visibilityStore := mgr.(*pinotVisibilityStore)
//
//			test.producerMockAffordance(mockProducer)
//
//			err := visibilityStore.RecordWorkflowExecutionStarted(context.Background(), test.request)
//			if test.expectedError != nil {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//			}
//		})
//	}
//}
