package resource

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShutdown(t *testing.T) {
	i := Impl{}
	assert.NotPanics(t, func() {
		i.Stop()
	})
}
