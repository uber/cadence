package pinotvisibility

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/.gen/go/indexer"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	metricsClientMocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/mocks"
	p "github.com/uber/cadence/common/persistence"
	pnt "github.com/uber/cadence/common/pinot"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
	"testing"
)

var (
	testStopwatch = metrics.NoopScope(metrics.PinotRecordWorkflowExecutionStartedScope).StartTimer(metrics.PinotLatency)
)

func TestMetricClientRecordWorkflowExecutionStarted(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.RecordWorkflowExecutionStartedRequest{
		WorkflowTypeName: "errorWorkflowTypeName",
		Memo: &types.Memo{
			map[string][]byte{
				"CustomStringField": []byte("test string"),
			},
		},
	}

	request := &p.RecordWorkflowExecutionStartedRequest{
		WorkflowTypeName: "wtn",
		Memo: &types.Memo{
			map[string][]byte{
				"CustomStringField": []byte("test string"),
			},
		},
	}

	tests := map[string]struct {
		request       *p.RecordWorkflowExecutionStartedRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: &types.BadRequestError{},
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(&types.BadRequestError{}).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.RecordWorkflowExecutionStarted(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.RecordWorkflowExecutionStarted(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientRecordWorkflowExecutionClosed(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.RecordWorkflowExecutionClosedRequest{
		WorkflowTypeName: "errorWorkflowTypeName",
		Memo: &types.Memo{
			map[string][]byte{
				"CustomStringField": []byte("test string"),
			},
		},
	}

	request := &p.RecordWorkflowExecutionClosedRequest{
		WorkflowTypeName: "wtn",
		Memo: &types.Memo{
			map[string][]byte{
				"CustomStringField": []byte("test string"),
			},
		},
	}

	tests := map[string]struct {
		request       *p.RecordWorkflowExecutionClosedRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: &types.ServiceBusyError{},
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(&types.ServiceBusyError{}).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.RecordWorkflowExecutionClosed(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.RecordWorkflowExecutionClosed(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientRecordWorkflowExecutionUninitialized(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.RecordWorkflowExecutionUninitializedRequest{
		WorkflowTypeName: "errorWorkflowTypeName",
	}

	request := &p.RecordWorkflowExecutionUninitializedRequest{
		WorkflowTypeName: "wtn",
	}

	tests := map[string]struct {
		request       *p.RecordWorkflowExecutionUninitializedRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(fmt.Errorf("error")).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.RecordWorkflowExecutionUninitialized(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.RecordWorkflowExecutionUninitialized(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientUpsertWorkflowExecution(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.UpsertWorkflowExecutionRequest{
		WorkflowTypeName: "errorWorkflowTypeName",
	}

	request := &p.UpsertWorkflowExecutionRequest{
		WorkflowTypeName: "wtn",
	}

	tests := map[string]struct {
		request       *p.UpsertWorkflowExecutionRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(fmt.Errorf("error")).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.UpsertWorkflowExecution(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.UpsertWorkflowExecution(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListOpenWorkflowExecutions(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsRequest{
		Domain:        DomainID,
		NextPageToken: []byte("error"),
	}

	request := &p.ListWorkflowExecutionsRequest{
		Domain: DomainID,
	}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListOpenWorkflowExecutions(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListOpenWorkflowExecutions(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListClosedWorkflowExecutions(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsRequest{
		Domain:        DomainID,
		NextPageToken: []byte("error"),
	}

	request := &p.ListWorkflowExecutionsRequest{
		Domain: DomainID,
	}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListClosedWorkflowExecutions(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListClosedWorkflowExecutions(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListOpenWorkflowExecutionsByType(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: p.ListWorkflowExecutionsRequest{
			Domain:        DomainID,
			NextPageToken: []byte("error"),
		},
		WorkflowTypeName: "",
	}

	request := &p.ListWorkflowExecutionsByTypeRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByTypeRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListOpenWorkflowExecutionsByType(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListOpenWorkflowExecutionsByType(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListClosedWorkflowExecutionsByType(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: p.ListWorkflowExecutionsRequest{
			Domain:        DomainID,
			NextPageToken: []byte("error"),
		},
		WorkflowTypeName: "",
	}

	request := &p.ListWorkflowExecutionsByTypeRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByTypeRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListClosedWorkflowExecutionsByType(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListClosedWorkflowExecutionsByType(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListOpenWorkflowExecutionsByWorkflowID(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: p.ListWorkflowExecutionsRequest{
			Domain:        DomainID,
			NextPageToken: []byte("error"),
		},
	}

	request := &p.ListWorkflowExecutionsByWorkflowIDRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByWorkflowIDRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListOpenWorkflowExecutionsByWorkflowID(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListOpenWorkflowExecutionsByWorkflowID(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListClosedWorkflowExecutionsByWorkflowID(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: p.ListWorkflowExecutionsRequest{
			Domain:        DomainID,
			NextPageToken: []byte("error"),
		},
	}

	request := &p.ListWorkflowExecutionsByWorkflowIDRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByWorkflowIDRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListClosedWorkflowExecutionsByWorkflowID(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListClosedWorkflowExecutionsByWorkflowID(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListClosedWorkflowExecutionsByStatus(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.ListClosedWorkflowExecutionsByStatusRequest{
		ListWorkflowExecutionsRequest: p.ListWorkflowExecutionsRequest{
			Domain:        DomainID,
			NextPageToken: []byte("error"),
		},
	}

	request := &p.ListClosedWorkflowExecutionsByStatusRequest{}

	tests := map[string]struct {
		request       *p.ListClosedWorkflowExecutionsByStatusRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListClosedWorkflowExecutionsByStatus(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListClosedWorkflowExecutionsByStatus(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientGetClosedWorkflowExecution(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.GetClosedWorkflowExecutionRequest{}

	request := &p.GetClosedWorkflowExecutionRequest{}

	tests := map[string]struct {
		request       *p.GetClosedWorkflowExecutionRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(&pnt.SearchResponse{
					Executions: []*p.InternalVisibilityWorkflowExecutionInfo{
						{
							DomainID: DomainID,
						},
					},
				}, fmt.Errorf("error")).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.GetClosedWorkflowExecution(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(&pnt.SearchResponse{
					Executions: []*p.InternalVisibilityWorkflowExecutionInfo{
						{
							DomainID: DomainID,
						},
					},
				}, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.GetClosedWorkflowExecution(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientListWorkflowExecutions(t *testing.T) {
	errorRequest := &p.ListWorkflowExecutionsByQueryRequest{
		NextPageToken: []byte("error"),
	}
	request := &p.ListWorkflowExecutionsByQueryRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByQueryRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ListWorkflowExecutions(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ListWorkflowExecutions(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientScanWorkflowExecutions(t *testing.T) {
	errorRequest := &p.ListWorkflowExecutionsByQueryRequest{
		NextPageToken: []byte("error"),
	}
	request := &p.ListWorkflowExecutionsByQueryRequest{}

	tests := map[string]struct {
		request       *p.ListWorkflowExecutionsByQueryRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.ScanWorkflowExecutions(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().Search(gomock.Any()).Return(nil, nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.ScanWorkflowExecutions(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientCountWorkflowExecutions(t *testing.T) {
	errorRequest := &p.CountWorkflowExecutionsRequest{}
	request := &p.CountWorkflowExecutionsRequest{}

	tests := map[string]struct {
		request       *p.CountWorkflowExecutionsRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().CountByQuery(gomock.Any()).Return(int64(0), fmt.Errorf("error")).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				_, err := metricsClient.CountWorkflowExecutions(context.Background(), test.request)
				assert.Error(t, err)
			} else {
				mockPinotClient.EXPECT().GetTableName().Return(testTableName).Times(1)
				mockPinotClient.EXPECT().CountByQuery(gomock.Any()).Return(int64(1), nil).Times(1)
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				_, err := metricsClient.CountWorkflowExecutions(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientDeleteWorkflowExecution(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.VisibilityDeleteWorkflowExecutionRequest{}

	request := &p.VisibilityDeleteWorkflowExecutionRequest{}

	tests := map[string]struct {
		request       *p.VisibilityDeleteWorkflowExecutionRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(fmt.Errorf("error")).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.DeleteWorkflowExecution(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.DeleteWorkflowExecution(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientDeleteUninitializedWorkflowExecution(t *testing.T) {
	// test non-empty request fields match
	errorRequest := &p.VisibilityDeleteWorkflowExecutionRequest{}

	request := &p.VisibilityDeleteWorkflowExecutionRequest{}

	tests := map[string]struct {
		request       *p.VisibilityDeleteWorkflowExecutionRequest
		expectedError error
	}{
		"Case1: error case": {
			request:       errorRequest,
			expectedError: fmt.Errorf("error"),
		},
		"Case2: normal case": {
			request:       request,
			expectedError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock clients
			ctrl := gomock.NewController(t)
			mockPinotClient := pnt.NewMockGenericClient(ctrl)
			mockProducer := &mocks.KafkaProducer{}
			mockMetricClient := &metricsClientMocks.Client{}
			mockScope := &metricsClientMocks.Scope{}

			// create metricClient
			logger := testlogger.New(t)
			mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
				ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
				ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
			}, mockProducer, testlogger.New(t))
			visibilityStore := mgr.(*pinotVisibilityStore)
			pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
			visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
			metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

			// mock behaviors
			mockMetricClient.On("Scope", mock.Anything, mock.Anything).Return(mockScope).Once()
			mockScope.On("StartTimer", mock.Anything, mock.Anything).Return(testStopwatch).Once()

			if test.expectedError != nil {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(fmt.Errorf("error")).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
				err := metricsClient.DeleteUninitializedWorkflowExecution(context.Background(), test.request)
				assert.Equal(t, err, test.expectedError)
			} else {
				mockProducer.On("Publish", mock.Anything, mock.MatchedBy(func(input *indexer.PinotMessage) bool {
					return true
				})).Return(nil).Once()
				mockScope.On("IncCounter", mock.Anything, mock.Anything, mock.Anything).Return().Once()
				err := metricsClient.DeleteUninitializedWorkflowExecution(context.Background(), test.request)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricClientClose(t *testing.T) {
	// create mock clients
	ctrl := gomock.NewController(t)
	mockPinotClient := pnt.NewMockGenericClient(ctrl)
	mockProducer := &mocks.KafkaProducer{}
	mockMetricClient := &metricsClientMocks.Client{}

	// create metricClient
	logger := testlogger.New(t)
	mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
		ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
		ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
	}, mockProducer, testlogger.New(t))
	visibilityStore := mgr.(*pinotVisibilityStore)
	pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
	visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
	metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

	assert.NotPanics(t, func() {
		metricsClient.Close()

	})
}

func TestMetricClientGetName(t *testing.T) {
	// create mock clients
	ctrl := gomock.NewController(t)
	mockPinotClient := pnt.NewMockGenericClient(ctrl)
	mockProducer := &mocks.KafkaProducer{}
	mockMetricClient := &metricsClientMocks.Client{}

	// create metricClient
	logger := testlogger.New(t)
	mgr := NewPinotVisibilityStore(mockPinotClient, &service.Config{
		ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
		ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(3),
	}, mockProducer, testlogger.New(t))
	visibilityStore := mgr.(*pinotVisibilityStore)
	pinotVisibilityManager := p.NewVisibilityManagerImpl(visibilityStore, logger)
	visibilityMgr := NewPinotVisibilityMetricsClient(pinotVisibilityManager, mockMetricClient, logger)
	metricsClient := visibilityMgr.(*pinotVisibilityMetricsClient)

	assert.NotPanics(t, func() {
		metricsClient.GetName()

	})
}
