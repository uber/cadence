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

package rpc

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/peer/direct"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/yarpcerrors"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
)

var (
	noOpSubscriberInstance = &noOpSubscriber{}
)

// directPeerChooser is a peer.Chooser that chooses a peer based on the shard key.
// Peers are managed by the peerList and peers are reused across multiple requests.
type directPeerChooser struct {
	serviceName          string
	logger               log.Logger
	t                    peer.Transport
	enableConnRetainMode dynamicconfig.BoolPropertyFn
	legacyChooser        peer.Chooser
	legacyChooserErr     error
	mu                   sync.RWMutex
	peers                map[string]peer.Peer
}

func newDirectChooser(serviceName string, t peer.Transport, logger log.Logger, enableConnRetainMode dynamicconfig.BoolPropertyFn) *directPeerChooser {
	return &directPeerChooser{
		serviceName:          serviceName,
		logger:               logger,
		t:                    t,
		enableConnRetainMode: enableConnRetainMode,
	}
}

// Start statisfies the peer.Chooser interface.
func (g *directPeerChooser) Start() error {
	c, ok := g.getLegacyChooser()
	if ok {
		return c.Start()
	}

	return nil // no-op
}

// Stop statisfies the peer.Chooser interface.
func (g *directPeerChooser) Stop() error {
	c, ok := g.getLegacyChooser()
	if ok {
		return c.Stop()
	}

	return nil // no-op
}

// IsRunning statisfies the peer.Chooser interface.
func (g *directPeerChooser) IsRunning() bool {
	c, ok := g.getLegacyChooser()
	if ok {
		return c.IsRunning()
	}

	return true // no-op
}

// Choose returns an existing peer for the shard key.
// ShardKey is {host}:{port} of the peer. It could be tchannel or grpc address.
func (g *directPeerChooser) Choose(ctx context.Context, req *transport.Request) (peer peer.Peer, onFinish func(error), err error) {
	if g.enableConnRetainMode != nil && !g.enableConnRetainMode() {
		return g.chooseFromLegacyDirectPeerChooser(ctx, req)
	}

	if req.ShardKey == "" {
		return nil, nil, yarpcerrors.InvalidArgumentErrorf("chooser requires ShardKey to be non-empty")
	}

	g.mu.RLock()
	p, ok := g.peers[req.ShardKey]
	if ok {
		g.mu.RUnlock()
		return p, func(error) {}, nil
	}
	g.mu.RUnlock()

	// peer is not cached, add new peer
	p, err = g.addPeer(req.ShardKey)
	if err != nil {
		return nil, nil, yarpcerrors.InternalErrorf("failed to add peer for shard key %v, err: %v", req.ShardKey, err)
	}

	return p, func(error) {}, nil
}

func (g *directPeerChooser) UpdatePeers(members []membership.HostInfo) {
	g.logger.Debug("direct peer chooser got a membership update", tag.Counter(len(members)))

	// Create a map of valid peer addresses given members list. This is used to remove peers that are not in the members list.
	// Actual yarpc peers are not created for the members here. They are created lazily when a request comes in.
	validPeerAddresses := make(map[string]bool)
	for _, member := range members {
		for _, portName := range []string{membership.PortTchannel, membership.PortGRPC} {
			addr, err := member.GetNamedAddress(portName)
			if err != nil {
				g.logger.Error(fmt.Sprintf("failed to get %s address of member", portName), tag.Error(err), tag.Address(member.GetAddress()))
				continue
			}
			validPeerAddresses[addr] = true
		}
	}

	// Take a copy of the current peers to avoid keeping write lock while removing all peers.
	peers := make(map[string]bool)
	g.mu.RLock()
	for addr := range g.peers {
		peers[addr] = true
	}
	g.mu.RUnlock()

	for addr := range peers {
		if !validPeerAddresses[addr] {
			g.removePeer(addr)
		}
	}

	// todo: emit gauge metric for peers
}

func (g *directPeerChooser) removePeer(addr string) {
	// TODO: is ReleasePeer thread safe? If not we need to hold the write lock while releasing the peer.
	if err := g.t.ReleasePeer(g.peers[addr], noOpSubscriberInstance); err != nil {
		g.logger.Error("failed to release peer", tag.Error(err), tag.Address(addr))
	}

	g.mu.Lock()
	delete(g.peers, addr)
	g.mu.Unlock()
	g.logger.Debug("removed peer from direct peer chooser", tag.Address(addr))
	// TODO: emit count metric
}

func (g *directPeerChooser) addPeer(addr string) (peer.Peer, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if p, ok := g.peers[addr]; ok {
		return p, nil
	}

	p, err := g.t.RetainPeer(hostport.Identify(addr), noOpSubscriberInstance)
	if err != nil {
		return nil, err
	}
	g.peers[addr] = p
	g.logger.Debug("added peer to direct peer chooser", tag.Address(addr))
	// TODO: emit count metric
	return p, nil
}

func (g *directPeerChooser) chooseFromLegacyDirectPeerChooser(ctx context.Context, req *transport.Request) (peer.Peer, func(error), error) {
	c, ok := g.getLegacyChooser()
	if !ok {
		return nil, nil, yarpcerrors.InternalErrorf("failed to get legacy direct peer chooser")
	}

	return c.Choose(ctx, req)
}

func (g *directPeerChooser) getLegacyChooser() (peer.Chooser, bool) {
	g.mu.RLock()

	if g.legacyChooser != nil {
		// Legacy chooser already created, return it
		g.mu.RUnlock()
		return g.legacyChooser, true
	}

	if g.legacyChooserErr != nil {
		// There was an error creating the legacy chooser, return false
		g.mu.RUnlock()
		return nil, false
	}

	g.mu.RUnlock()

	g.mu.Lock()
	g.legacyChooser, g.legacyChooserErr = direct.New(direct.Configuration{}, g.t)
	g.mu.Unlock()

	if g.legacyChooserErr != nil {
		g.logger.Error("failed to create legacy direct peer chooser", tag.Error(g.legacyChooserErr))
		return nil, false
	}

	return g.legacyChooser, true
}

// noOpSubscriber is a peer.Subscriber that does nothing.
type noOpSubscriber struct{}

func (*noOpSubscriber) NotifyStatusChanged(peer.Identifier) {}
