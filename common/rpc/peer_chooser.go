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
	"time"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/peer/direct"
	"go.uber.org/yarpc/peer/roundrobin"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
)

const defaultDNSRefreshInterval = time.Second * 10

type (
	PeerChooserOptions struct {
		// Address is used by dns peer chooser
		Address string
	}
	PeerChooserFactory interface {
		CreatePeerChooser(transport peer.Transport, opts PeerChooserOptions) (peer.Chooser, error)

		// PeerChooserFactory is a daemon because it may run background processes. CreatePeerChooser is called before Start.
		common.Daemon
	}

	dnsPeerChooserFactory struct {
		interval time.Duration
		logger   log.Logger
	}

	directPeerChooserFactory struct {
		logger log.Logger
	}
)

func NewDNSPeerChooserFactory(interval time.Duration, logger log.Logger) PeerChooserFactory {
	if interval <= 0 {
		interval = defaultDNSRefreshInterval
	}

	return &dnsPeerChooserFactory{interval, logger}
}

func (f *dnsPeerChooserFactory) CreatePeerChooser(transport peer.Transport, opts PeerChooserOptions) (peer.Chooser, error) {
	peerList := roundrobin.New(transport)
	peerListUpdater, err := newDNSUpdater(peerList, opts.Address, f.interval, f.logger)
	if err != nil {
		return nil, err
	}
	peerListUpdater.Start()
	return peerList, nil
}

func (f *dnsPeerChooserFactory) Start() {}
func (f *dnsPeerChooserFactory) Stop()  {}

func NewDirectPeerChooserFactory(logger log.Logger) PeerChooserFactory {
	return &directPeerChooserFactory{
		logger: logger,
	}
}

func (f *directPeerChooserFactory) CreatePeerChooser(transport peer.Transport, opts PeerChooserOptions) (peer.Chooser, error) {
	return direct.New(direct.Configuration{}, transport)
}

func (f *directPeerChooserFactory) Start() {
	// TODO: subscribe to membership changes
}

func (f *directPeerChooserFactory) Stop() {
	// stop all connections and background goroutine
}
