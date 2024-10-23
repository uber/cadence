package nosql

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/log"
	"testing"
)

func TestNewNoSQLVisibilityStore(t *testing.T) {
	registerCassandraMock(t)
	cfg := getValidShardedNoSQLConfig()

	store, err := newNoSQLVisibilityStore(false, cfg, log.NewNoop(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}
