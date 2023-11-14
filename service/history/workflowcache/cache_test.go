package workflowcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testDomainID   = "B59344B2-4166-462D-9CBD-22B25D2A7B1B"
	testWorkflowID = "8ED9219B-36A2-4FD0-B9EA-6298A0F2ED1A"
)

func Test_wfCache_Allow(t *testing.T) {
	c := New(
		Params{
			TTL:      60 * time.Second,
			MaxCount: 10_000,
			MaxRPS: func() float64 {
				return 2.0
			},
		})

	// The burst size is set to the MaxRPS, so we can call Allow() twice without delay
	assert.True(t, c.Allow(testDomainID, testWorkflowID))
	assert.True(t, c.Allow(testDomainID, testWorkflowID))

	// The next token will be added after 1/MaxRPS = 500ms, so if we only wait 400ms, the next call will fail
	time.Sleep(400 * time.Millisecond)
	assert.False(t, c.Allow(testDomainID, testWorkflowID))

	// After 500ms, the next token will be added, so waiting an additional 200ms, the next call will succeed
	time.Sleep(200 * time.Millisecond)
	assert.True(t, c.Allow(testDomainID, testWorkflowID))
}
