package cassandra

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"testing"
	"time"
)

func Test_toGoCqlConfig(t *testing.T) {

	tests := []struct {
		name    string
		cfg     *config.NoSQL
		want    gocql.ClusterConfig
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"empty config will be filled with defaults",
			&config.NoSQL{},
			gocql.ClusterConfig{
				Hosts:             "127.0.0.1",
				Port:              9042,
				ProtoVersion:      4,
				Timeout:           time.Second * 10,
				Consistency:       gocql.LocalQuorum,
				SerialConsistency: gocql.LocalSerial,
				ConnectTimeout:    time.Second * 2,
			},
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toGoCqlConfig(tt.cfg)
			if !tt.wantErr(t, err, fmt.Sprintf("toGoCqlConfig(%v)", tt.cfg)) {
				return
			}
			assert.Equalf(t, tt.want, got, "toGoCqlConfig(%v)", tt.cfg)
		})
	}
}
