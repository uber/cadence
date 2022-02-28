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

	host := HostInfo{}

	belongs, err := host.Belongs("127.1")
	assert.False(t, belongs, "invalid host info data should result to false")
	assert.Error(t, err)

	host2 := NewDetailedHostInfo("127.0.0.1:123", "dummy", PortMap{})
	belongs, err = host2.Belongs("127.0.0.1:123")
	assert.True(t, belongs, "match on address, port map might be empty")
	assert.NoError(t, err)

	host3 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host3.Belongs("127.0.0.1:3333")
	assert.True(t, belongs, "portmap should be checked")
	assert.NoError(t, err)

	host4 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host4.Belongs("127.0.0.2:3333")
	assert.False(t, belongs, "different IP, will result in false")
	assert.NoError(t, err)

	host5 := NewDetailedHostInfo("127.0.0.1:1234", "dummy", PortMap{PortGRPC: 3333})
	belongs, err = host5.Belongs("127.0.0.1:3334")
	assert.False(t, belongs, "portmap has no such port, should return empty without an error")
	assert.NoError(t, err)
}
