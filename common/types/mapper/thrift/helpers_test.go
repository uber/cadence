package thrift

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeToNano(t *testing.T) {
	unixTime := time.Unix(1, 1)
	result := timeToNano(&unixTime)
	assert.Equal(t, int64(1000000001), *result)
}

func TestTimeToNanoNil(t *testing.T) {
	result := timeToNano(nil)
	assert.Nil(t, result)
}
