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
func (*directPeerChooser) Start() error {
	return nil // no-op
}

// Stop statisfies the peer.Chooser interface.
func (*directPeerChooser) Stop() error {
	return nil // no-op
}

// IsRunning statisfies the peer.Chooser interface.
func (*directPeerChooser) IsRunning() bool {
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

func (g *directPeerChooser) UpdatePeers([]membership.HostInfo) {
	// TODO: implement
}

func (g *directPeerChooser) chooseFromLegacyDirectPeerChooser(ctx context.Context, req *transport.Request) (peer.Peer, func(error), error) {
	g.mu.RLock()

	if g.legacyChooser != nil {
		g.mu.RUnlock()
		return g.legacyChooser.Choose(ctx, req)
	}

	g.mu.RUnlock()

	var err error
	g.mu.Lock()
	g.legacyChooser, err = direct.New(direct.Configuration{}, g.t)
	g.mu.Unlock()

	if err != nil {
		return nil, nil, err
	}

	return g.legacyChooser.Choose(ctx, req)
}
