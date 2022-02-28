package membership

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBelongs(t *testing.T) {

	host := HostInfo{
		addr:     "",
		ip:       "",
		identity: "",
		portMap:  nil,
	}

	belongs, err := host.Belongs("127.1")
	assert.False(t, belongs)
	assert.Error(t, err)

	// based on address, without port map
	host2 := NewDetailedHostInfo("127.0.0.1:123", "dummy", PortMap{})
	belongs, err = host2.Belongs("127.0.0.1:123")
	assert.True(t, belongs)
	assert.NoError(t, err)

	// based on port map
	host3 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host3.Belongs("127.0.0.1:3333")
	assert.True(t, belongs)
	assert.NoError(t, err)

	// different ip will result in false
	host4 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host4.Belongs("127.0.0.2:3333")
	assert.False(t, belongs)
	assert.NoError(t, err)

	// port is not found in portmap
	host5 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host5.Belongs("127.0.0.1:3334")
	assert.False(t, belongs)
	assert.NoError(t, err)
}
