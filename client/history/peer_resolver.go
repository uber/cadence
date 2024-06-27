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
	"fmt"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/service/history/lookup"
)

// PeerResolver is used to resolve history peers.
// Those are deployed instances of Cadence history services that participate in the cluster ring.
// The resulting peer is simply an address of form ip:port where RPC calls can be routed to.
//
//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination peer_resolver_mock.go -package history github.com/uber/cadence/client/history PeerResolver
type PeerResolver interface {
	FromWorkflowID(workflowID string) (string, error)
	FromDomainID(domainID string) (string, error)
	FromShardID(shardID int) (string, error)
	FromHostAddress(hostAddress string) (string, error)
	GetAllPeers() ([]string, error)

	// GlobalRatelimitPeers partitions the ratelimit keys into map[yarpc peer][]limits_for_peer
	GlobalRatelimitPeers(ratelimits []string) (ratelimitsByPeer map[Peer][]string, err error)
}

// Peer is used to mark a string as the routing information to a peer process.
//
// This is essentially the host:port address of the peer to be contacted,
// but it is meant to be treated as an opaque blob until given to yarpc
// via ToYarpcShardKey.
type Peer string

func (s Peer) ToYarpcShardKey() yarpc.CallOption {
	return yarpc.WithShardKey(string(s))
}

type peerResolver struct {
	numberOfShards int
	resolver       membership.Resolver
	namedPort      string // grpc or tchannel, depends on yarpc configuration
}

// NewPeerResolver creates a new history peer resolver.
func NewPeerResolver(numberOfShards int, resolver membership.Resolver, namedPort string) PeerResolver {
	return peerResolver{
		numberOfShards: numberOfShards,
		resolver:       resolver,
		namedPort:      namedPort,
	}
}

// FromWorkflowID resolves the history peer responsible for a given workflowID.
// WorkflowID is converted to logical shardID using a consistent hash function.
// FromShardID is used for further resolving.
func (pr peerResolver) FromWorkflowID(workflowID string) (string, error) {
	shardID := common.WorkflowIDToHistoryShard(workflowID, pr.numberOfShards)
	return pr.FromShardID(shardID)
}

// FromDomainID resolves the history peer responsible for a given domainID.
// DomainID is converted to logical shardID using a consistent hash function.
// FromShardID is used for further resolving.
func (pr peerResolver) FromDomainID(domainID string) (string, error) {
	shardID := common.DomainIDToHistoryShard(domainID, pr.numberOfShards)
	return pr.FromShardID(shardID)
}

// FromShardID resolves the history peer responsible for a given logical shardID.
// It uses our membership provider to lookup which instance currently owns the given shard.
// FromHostAddress is used for further resolving.
func (pr peerResolver) FromShardID(shardID int) (string, error) {
	host, err := lookup.HistoryServerByShardID(pr.resolver, shardID)
	if err != nil {
		return "", common.ToServiceTransientError(err)
	}
	peer, err := host.GetNamedAddress(pr.namedPort)
	return peer, common.ToServiceTransientError(err)
}

// FromHostAddress resolves the final history peer responsible for the given host address.
// The address is formed by adding port for specified transport
func (pr peerResolver) FromHostAddress(hostAddress string) (string, error) {
	host, err := pr.resolver.LookupByAddress(service.History, hostAddress)
	if err != nil {
		return "", common.ToServiceTransientError(err)
	}
	peer, err := host.GetNamedAddress(pr.namedPort)
	return peer, common.ToServiceTransientError(err)
}

// GetAllPeers returns all history service peers in the cluster ring.
func (pr peerResolver) GetAllPeers() ([]string, error) {
	hosts, err := pr.resolver.Members(service.History)
	if err != nil {
		return nil, common.ToServiceTransientError(err)
	}
	peers := make([]string, 0, len(hosts))
	for _, host := range hosts {
		peer, err := host.GetNamedAddress(pr.namedPort)
		if err != nil {
			return nil, common.ToServiceTransientError(err)
		}
		peers = append(peers, peer)
	}
	return peers, nil
}

func (pr peerResolver) GlobalRatelimitPeers(ratelimits []string) (map[Peer][]string, error) {
	// History was selected simply because it already has a ring and an internal-only API.
	// Any service should be fine, it just needs to be shared by both ends of the system.
	hosts, err := pr.resolver.Members(service.History)
	if err != nil {
		return nil, fmt.Errorf("unable to get history peers: %w", err)
	}
	if len(hosts) == 0 {
		// can occur when shutting down the only instance because it calls EvictSelf ASAP.
		// this is common in dev, but otherwise *probably* should not be possible in a healthy system.
		return nil, fmt.Errorf("unable to get history peers: no peers available")
	}

	results := make(map[Peer][]string, len(hosts))
	initialCapacity := len(ratelimits) / len(hosts)
	// add a small buffer to reduce copies, as this is only an estimate
	initialCapacity += initialCapacity / 10
	// but don't use zero, that'd be pointless
	if initialCapacity < 1 {
		initialCapacity = 1
	}

	for _, r := range ratelimits {
		// figure out the destination for this ratelimit
		host, err := pr.resolver.Lookup(service.History, r)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to find host for ratelimit key %q: %w", r, err)
		}
		peer, err := host.GetNamedAddress(pr.namedPort)
		if err != nil {
			// AFAICT should not happen unless the peer data is incomplete / the ring has no port info.
			// retry may work?  eventually?
			return nil, fmt.Errorf(
				"unable to get address from host: %s, %w",
				host.String(), // HostInfo has a readable String(), it's good enough
				err,
			)
		}

		// add to the peer's partition
		current := results[Peer(peer)]
		if len(current) == 0 {
			current = make([]string, 0, initialCapacity)
		}
		current = append(current, r)
		results[Peer(peer)] = current
	}
	return results, nil
}
