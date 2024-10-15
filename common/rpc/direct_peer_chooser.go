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
	"sync"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/peer/direct"
	"go.uber.org/yarpc/yarpcerrors"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
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
func (g *directPeerChooser) Choose(ctx context.Context, req *transport.Request) (peer peer.Peer, onFinish func(error), err error) {
	if g.enableConnRetainMode != nil && !g.enableConnRetainMode() {
		return g.chooseFromLegacyDirectPeerChooser(ctx, req)
	}

	if req.ShardKey == "" {
		return nil, nil, yarpcerrors.InvalidArgumentErrorf("chooser requires ShardKey to be non-empty")
	}

	// TODO: implement connection retain mode
	return nil, nil, yarpcerrors.UnimplementedErrorf("direct peer chooser conn retain mode unimplemented")
}

func (g *directPeerChooser) UpdatePeers(members []membership.HostInfo) {
	// TODO: implement
	g.logger.Debug("direct peer chooser got a membership update", tag.Counter(len(members)))
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
