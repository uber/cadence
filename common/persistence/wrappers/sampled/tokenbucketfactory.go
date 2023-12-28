package sampled

import (
	"sync"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/tokenbucket"
)

type RateLimiterFactoryFunc func(timeSource clock.TimeSource, numOfPriority int, qpsConfig dynamicconfig.IntPropertyFnWithDomainFilter) RateLimiterFactory

type RateLimiterFactory interface {
	GetRateLimiter(domain string) tokenbucket.PriorityTokenBucket
}

type domainToBucketMap struct {
	sync.RWMutex
	timeSource    clock.TimeSource
	qpsConfig     dynamicconfig.IntPropertyFnWithDomainFilter
	numOfPriority int
	mappings      map[string]tokenbucket.PriorityTokenBucket
}

// NewDomainToBucketMap returns a rate limiter factory.
func NewDomainToBucketMap(timeSource clock.TimeSource, numOfPriority int, qpsConfig dynamicconfig.IntPropertyFnWithDomainFilter) RateLimiterFactory {
	return &domainToBucketMap{
		timeSource:    timeSource,
		qpsConfig:     qpsConfig,
		numOfPriority: numOfPriority,
		mappings:      make(map[string]tokenbucket.PriorityTokenBucket),
	}
}

func (m *domainToBucketMap) GetRateLimiter(domain string) tokenbucket.PriorityTokenBucket {
	m.RLock()
	rateLimiter, exist := m.mappings[domain]
	m.RUnlock()

	if exist {
		return rateLimiter
	}

	m.Lock()
	if rateLimiter, ok := m.mappings[domain]; ok { // read again to ensure no duplicate create
		m.Unlock()
		return rateLimiter
	}
	rateLimiter = tokenbucket.NewFullPriorityTokenBucket(m.numOfPriority, m.qpsConfig(domain), m.timeSource)
	m.mappings[domain] = rateLimiter
	m.Unlock()
	return rateLimiter
}
