// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package matching

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
)

func TestPeerResolver(t *testing.T) {
	controller := gomock.NewController(t)
	serviceResolver := membership.NewMockResolver(controller)
	serviceResolver.EXPECT().Lookup(service.Matching, "taskListA").Return(membership.NewHostInfo("taskListA:thriftPort"), nil)
	serviceResolver.EXPECT().Lookup(service.Matching, "invalid").Return(membership.HostInfo{}, assert.AnError)
	serviceResolver.EXPECT().Members(service.Matching).Return([]membership.HostInfo{
		membership.NewHostInfo("taskListA:thriftPort"),
		membership.NewHostInfo("taskListB:thriftPort"),
	}, nil)

	r := NewPeerResolver(serviceResolver, fakeAddressMapper)

	peer, err := r.FromTaskList("taskListA")
	assert.NoError(t, err)
	assert.Equal(t, "taskListA:grpcPort", peer)

	_, err = r.FromTaskList("invalid")
	assert.Error(t, err)

	_, err = r.FromHostAddress("invalid address")
	assert.Error(t, err)

	peers, err := r.GetAllPeers()
	assert.NoError(t, err)
	assert.Equal(t, []string{"taskListA:grpcPort", "taskListB:grpcPort"}, peers)

	r = NewPeerResolver(nil, nil)
	peer, err = r.FromHostAddress("no mapper")
	assert.NoError(t, err)
	assert.Equal(t, "no mapper", peer)
}

func fakeAddressMapper(address string) (string, error) {
	switch address {
	case "taskListA:thriftPort":
		return "taskListA:grpcPort", nil
	case "taskListB:thriftPort":
		return "taskListB:grpcPort", nil
	}
	return "", assert.AnError
}
