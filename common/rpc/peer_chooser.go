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

	"github.com/uber/cadence/common/log"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/peer/roundrobin"
)

const defaultDNSRefreshInterval = time.Second * 10

type (
	PeerChooserFactory interface {
		CreatePeerChooser(transport peer.Transport, address string) (peer.Chooser, error)
	}
	dnsPeerChooserFactory struct {
		interval time.Duration
		logger   log.Logger
	}
)

func NewDNSPeerChooserFactory(interval time.Duration, logger log.Logger) *dnsPeerChooserFactory {
	if interval <= 0 {
		interval = defaultDNSRefreshInterval
	}

	return &dnsPeerChooserFactory{interval, logger}
}

func (f *dnsPeerChooserFactory) CreatePeerChooser(transport peer.Transport, address string) (peer.Chooser, error) {
	peerList := roundrobin.New(transport)
	peerListUpdater, err := newDNSUpdater(peerList, address, f.interval, f.logger)
	if err != nil {
		return nil, err
	}
	peerListUpdater.Start()
	return peerList, nil
}
