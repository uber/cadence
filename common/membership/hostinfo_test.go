// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
