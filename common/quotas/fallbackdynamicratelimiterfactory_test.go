package quotas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/dynamicconfig"
)

func TestNewFallbackDynamicRateLimiterFactory(t *testing.T) {
	factory := NewFallbackDynamicRateLimiterFactory(
		func(string) int { return 2 },
		func(opts ...dynamicconfig.FilterOption) int { return 100 },
	)

	limiter := factory.GetLimiter("TestDomainName")

	// The limiter should accept 2 requests per second
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, false, limiter.Allow())
}

func TestNewFallbackDynamicRateLimiterFactoryFallback(t *testing.T) {
	factory := NewFallbackDynamicRateLimiterFactory(
		func(string) int { return 0 },
		func(opts ...dynamicconfig.FilterOption) int { return 2 },
	)

	limiter := factory.GetLimiter("TestDomainName")

	// The limiter should accept 2 requests per second
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, true, limiter.Allow())
	assert.Equal(t, false, limiter.Allow())
}
