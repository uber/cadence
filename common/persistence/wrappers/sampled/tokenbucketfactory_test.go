package sampled

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
)

func TestDomainToBucketMap(t *testing.T) {
	mockedTime := clock.NewMockedTimeSource()
	factory := NewDomainToBucketMap(mockedTime, 1, dynamicconfig.GetIntPropertyFilteredByDomain(1))

	// Test that the factory returns the same bucket for the same domain
	bucket1 := factory.GetRateLimiter("domain1")
	bucket2 := factory.GetRateLimiter("domain1")
	assert.Equal(t, bucket1, bucket2, "domain bucket should return the same bucket for the same domain")
}
