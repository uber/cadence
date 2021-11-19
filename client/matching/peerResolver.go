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
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
)

// PeerResolver is used to resolve matching peers.
// Those are deployed instances of Cadence matching services that participate in the cluster ring.
// The resulting peer is simply an address of form ip:port where RPC calls can be routed to.
type PeerResolver struct {
	resolver      membership.Resolver
	addressMapper AddressMapperFn
}

type AddressMapperFn func(string) (string, error)

// NewPeerResolver creates a new matching peer resolver.
func NewPeerResolver(membership membership.Resolver, addressMapper AddressMapperFn) PeerResolver {
	return PeerResolver{membership, addressMapper}
}

// FromTaskList resolves the matching peer responsible for the given task list name.
// It uses our membership provider to lookup which instance currently owns the given task list.
// FromHostAddress is used for further resolving.
func (pr PeerResolver) FromTaskList(taskListName string) (string, error) {
	host, err := pr.resolver.Lookup(service.Matching, taskListName)
	if err != nil {
		return "", err
	}

	return pr.FromHostAddress(host.GetAddress())
}

// GetAllPeers returns all matching service peers in the cluster ring.
func (pr PeerResolver) GetAllPeers() ([]string, error) {
	hosts, err := pr.resolver.Members(service.Matching)
	if err != nil {
		return nil, err
	}
	peers := make([]string, 0, len(hosts))
	for _, host := range hosts {
		peer, err := pr.FromHostAddress(host.GetAddress())
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	return peers, nil
}

// FromHostAddress resolves the final matching peer responsible for the given host address.
// The address may be used as is, or processed with additional address mapper.
// In case of gRPC transport, the port within the address is replaced with gRPC port.
func (pr PeerResolver) FromHostAddress(hostAddress string) (string, error) {
	if pr.addressMapper == nil {
		return hostAddress, nil
	}
	return pr.addressMapper(hostAddress)
}
