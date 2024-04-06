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
	"errors"
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
	serviceResolver.EXPECT().Lookup(
		service.History, string(rune(common.DomainIDToHistoryShard("domainID", numShards)))).Return(
		membership.NewDetailedHostInfo(
			"domainHost:123",
			"domainHost_123",
			membership.PortMap{membership.PortTchannel: 1234}),
		nil)
	serviceResolver.EXPECT().Lookup(service.History, string(rune(common.WorkflowIDToHistoryShard("workflowID", numShards)))).Return(
		membership.NewDetailedHostInfo(
			"workflowHost:123",
			"workflow",
			membership.PortMap{membership.PortTchannel: 1235, membership.PortGRPC: 1666}), nil)

	serviceResolver.EXPECT().Lookup(service.History, string(rune(99))).Return(
		membership.NewDetailedHostInfo(
			"shardHost:123",
			"shard_123",
			membership.PortMap{membership.PortTchannel: 1235}),
		nil)

	serviceResolver.EXPECT().Lookup(service.History, string(rune(11))).Return(membership.HostInfo{}, assert.AnError)

	r := NewPeerResolver(numShards, serviceResolver, membership.PortTchannel)

	peer, err := r.FromDomainID("domainID")
	assert.NoError(t, err)
	assert.Equal(t, "domainHost:1234", peer)

	peer, err = r.FromWorkflowID("workflowID")
	assert.NoError(t, err)
	assert.Equal(t, "workflowHost:1235", peer)

	peer, err = r.FromShardID(99)
	assert.NoError(t, err)
	assert.Equal(t, "shardHost:1235", peer)

	_, err = r.FromShardID(11)
	assert.Error(t, err)

	t.Run("FromHostAddress", func(t *testing.T) {
		tests := []struct {
			name      string
			address   string
			mock      func(*membership.MockResolver)
			want      string
			wantError bool
		}{
			{
				name:    "success",
				address: "addressHost:123",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().LookupByAddress(service.History, "addressHost:123").Return(
						membership.NewDetailedHostInfo(
							"addressHost:123",
							"address",
							membership.PortMap{membership.PortTchannel: 1235, membership.PortGRPC: 1666}),
						nil,
					)
				},
				want: "addressHost:1235",
			},
			{
				name:    "invalid address",
				address: "invalid address",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().LookupByAddress(service.History, "invalid address").Return(
						membership.HostInfo{},
						errors.New("host not found"),
					)
				},
				wantError: true,
			},
			{
				name:    "fail on no port",
				address: "addressHost:123",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().LookupByAddress(service.History, "addressHost:123").Return(
						membership.NewDetailedHostInfo(
							"addressHost:123",
							"address",
							membership.PortMap{}),
						nil,
					)
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				controller := gomock.NewController(t)
				serviceResolver := membership.NewMockResolver(controller)
				tt.mock(serviceResolver)
				r := NewPeerResolver(numShards, serviceResolver, membership.PortTchannel)
				res, err := r.FromHostAddress(tt.address)
				if tt.wantError {
					assert.True(t, common.IsServiceTransientError(err))
				} else {
					assert.Equal(t, tt.want, res)
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("GetAllPeers", func(t *testing.T) {
		tests := []struct {
			name      string
			mock      func(*membership.MockResolver)
			want      []string
			wantError bool
		}{
			{
				name: "success",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().Members(service.History).Return(
						[]membership.HostInfo{
							membership.NewDetailedHostInfo(
								"host1:123",
								"address",
								membership.PortMap{membership.PortTchannel: 1235, membership.PortGRPC: 1666}),
							membership.NewDetailedHostInfo(
								"host2:123",
								"address",
								membership.PortMap{membership.PortTchannel: 1235, membership.PortGRPC: 1666}),
						},
						nil,
					)
				},
				want: []string{"host1:1235", "host2:1235"},
			},
			{
				name: "failed on peer resolve",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().Members(service.History).Return(nil, assert.AnError)

				},
				wantError: true,
			},
			{
				name: "failed on no port",
				mock: func(mr *membership.MockResolver) {
					mr.EXPECT().Members(service.History).Return(
						[]membership.HostInfo{
							membership.NewDetailedHostInfo(
								"host1:123",
								"address",
								membership.PortMap{membership.PortTchannel: 1235, membership.PortGRPC: 1666}),
							membership.NewDetailedHostInfo(
								"host2:123",
								"address",
								membership.PortMap{membership.PortGRPC: 1666}),
						},
						nil,
					)
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				controller := gomock.NewController(t)
				serviceResolver := membership.NewMockResolver(controller)
				tt.mock(serviceResolver)
				r := NewPeerResolver(numShards, serviceResolver, membership.PortTchannel)
				res, err := r.GetAllPeers()
				if tt.wantError {
					assert.True(t, common.IsServiceTransientError(err))
				} else {
					assert.ElementsMatch(t, tt.want, res)
					assert.NoError(t, err)
				}
			})
		}
	})
}
