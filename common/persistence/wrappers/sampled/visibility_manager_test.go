package sampled

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/tokenbucket"
)

func TestVisibilityManager_RecordWorkflowExecutionStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().RecordWorkflowExecutionStarted(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := m.RecordWorkflowExecutionStarted(context.Background(), &persistence.RecordWorkflowExecutionStartedRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "first call should succeed")

	err = m.RecordWorkflowExecutionStarted(context.Background(), &persistence.RecordWorkflowExecutionStartedRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "second call should succeed, but underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_RecordWorkflowExecutionClosed(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{1: 1}},
		}),
	})

	mockedManager.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := m.RecordWorkflowExecutionClosed(context.Background(), &persistence.RecordWorkflowExecutionClosedRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "first call should succeed")

	err = m.RecordWorkflowExecutionClosed(context.Background(), &persistence.RecordWorkflowExecutionClosedRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "second call should succeed, but underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_UpsertWorkflowExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().UpsertWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := m.UpsertWorkflowExecution(context.Background(), &persistence.UpsertWorkflowExecutionRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "first call should succeed")

	err = m.UpsertWorkflowExecution(context.Background(), &persistence.UpsertWorkflowExecutionRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "second call should succeed, but underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListOpenWorkflowExecutions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListOpenWorkflowExecutions(context.Background(), &persistence.ListWorkflowExecutionsRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListOpenWorkflowExecutions(context.Background(), &persistence.ListWorkflowExecutionsRequest{
		Domain: testDomain,
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListClosedWorkflowExecutions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListClosedWorkflowExecutions(context.Background(), &persistence.ListWorkflowExecutionsRequest{
		Domain: testDomain,
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListClosedWorkflowExecutions(context.Background(), &persistence.ListWorkflowExecutionsRequest{
		Domain: testDomain,
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListOpenWorkflowExecutionsByType(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListOpenWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListOpenWorkflowExecutionsByType(context.Background(), &persistence.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListOpenWorkflowExecutionsByType(context.Background(), &persistence.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListClosedWorkflowExecutionsByType(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListClosedWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListClosedWorkflowExecutionsByType(context.Background(), &persistence.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListClosedWorkflowExecutionsByType(context.Background(), &persistence.ListWorkflowExecutionsByTypeRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListOpenWorkflowExecutionsByWorkflowID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListOpenWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListOpenWorkflowExecutionsByWorkflowID(context.Background(), &persistence.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListOpenWorkflowExecutionsByWorkflowID(context.Background(), &persistence.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func TestVisibilityManager_ListClosedWorkflowExecutionsByWorkflowID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockedManager := persistence.NewMockVisibilityManager(ctrl)

	testDomain := "domain1"

	m := NewVisibilityManager(mockedManager, Params{
		Config:       &Config{},
		MetricClient: metrics.NewNoopMetricsClient(),
		Logger:       testlogger.New(t),
		TimeSource:   clock.NewMockedTimeSource(),
		RateLimiterFactoryFunc: rateLimiterStubFunc(map[string]tokenbucket.PriorityTokenBucket{
			testDomain: &tokenBucketFactoryStub{tokens: map[int]int{0: 1}},
		}),
	})

	mockedManager.EXPECT().ListClosedWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Times(1)

	_, err := m.ListClosedWorkflowExecutionsByWorkflowID(context.Background(), &persistence.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.NoError(t, err, "first call should succeed")

	_, err = m.ListClosedWorkflowExecutionsByWorkflowID(context.Background(), &persistence.ListWorkflowExecutionsByWorkflowIDRequest{
		ListWorkflowExecutionsRequest: persistence.ListWorkflowExecutionsRequest{
			Domain: testDomain,
		},
	})
	assert.Error(t, err, "second call should fail since underlying call should be blocked by rate limiter")
}

func rateLimiterStubFunc(domainData map[string]tokenbucket.PriorityTokenBucket) RateLimiterFactoryFunc {
	return func(timeSource clock.TimeSource, numOfPriority int, qpsConfig dynamicconfig.IntPropertyFnWithDomainFilter) RateLimiterFactory {
		return rateLimiterStub{domainData}
	}
}

type rateLimiterStub struct {
	data map[string]tokenbucket.PriorityTokenBucket
}

func (r rateLimiterStub) GetRateLimiter(domain string) tokenbucket.PriorityTokenBucket {
	return r.data[domain]
}

type tokenBucketFactoryStub struct {
	tokens map[int]int
}

func (t *tokenBucketFactoryStub) GetToken(priority, count int) (bool, time.Duration) {
	val := t.tokens[priority]
	if count > val {
		return false, time.Duration(0)
	}
	val -= count
	t.tokens[priority] = val
	return true, time.Duration(0)
}
