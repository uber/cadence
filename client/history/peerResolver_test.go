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

package history

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
)

func TestPeerResolver(t *testing.T) {
	numShards := 123
	controller := gomock.NewController(t)
	serviceResolver := membership.NewMockResolver(controller)
	serviceResolver.EXPECT().Lookup(service.History, string(rune(common.DomainIDToHistoryShard("domainID", numShards)))).Return(membership.NewHostInfo("domainHost:thriftPort"), nil)
	serviceResolver.EXPECT().Lookup(service.History, string(rune(common.WorkflowIDToHistoryShard("workflowID", numShards)))).Return(membership.NewHostInfo("workflowHost:thriftPort"), nil)
	serviceResolver.EXPECT().Lookup(service.History, string(rune(99))).Return(membership.NewHostInfo("shardHost:thriftPort"), nil)
	serviceResolver.EXPECT().Lookup(service.History, string(rune(11))).Return(membership.HostInfo{}, assert.AnError)

	r := NewPeerResolver(numShards, serviceResolver, fakeAddressMapper)

	peer, err := r.FromDomainID("domainID")
	assert.NoError(t, err)
	assert.Equal(t, "domainHost:grpcPort", peer)

	peer, err = r.FromWorkflowID("workflowID")
	assert.NoError(t, err)
	assert.Equal(t, "workflowHost:grpcPort", peer)

	peer, err = r.FromShardID(99)
	assert.NoError(t, err)
	assert.Equal(t, "shardHost:grpcPort", peer)

	_, err = r.FromShardID(11)
	assert.Error(t, err)

	_, err = r.FromHostAddress("invalid address")
	assert.Error(t, err)

	r = NewPeerResolver(numShards, nil, nil)
	peer, err = r.FromHostAddress("no mapper")
	assert.NoError(t, err)
	assert.Equal(t, "no mapper", peer)
}

func fakeAddressMapper(address string) (string, error) {
	switch address {
	case "domainHost:thriftPort":
		return "domainHost:grpcPort", nil
	case "workflowHost:thriftPort":
		return "workflowHost:grpcPort", nil
	case "shardHost:thriftPort":
		return "shardHost:grpcPort", nil
	}
	return "", assert.AnError
}
