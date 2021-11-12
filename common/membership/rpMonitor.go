// Copyright (c) 2017 Uber Technologies, Inc.
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

package membership

import (
	"fmt"

	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/swim"
	tcg "github.com/uber/tchannel-go"
	"go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/cadence/common/log"
)

type RingpopMonitor struct {
	status int32

	serviceName    string
	ringpopWrapper *RingpopWrapper
	rings          map[string]*ringpopServiceResolver
	logger         log.Logger
}

var _ Monitor = (*RingpopMonitor)(nil)

// NewRingpopMonitor builds a ringpop monitor conforming
// to the underlying configuration
func NewMonitor(
	config *RingpopConfig,
	channel tchannel.Channel,
	serviceName string,
	logger log.Logger,
) (*RingpopMonitor, error) {

	if err := config.validate(); err != nil {
		return nil, err
	}

	rp, err := ringpop.New(config.Name, ringpop.Channel(channel.(*tcg.Channel)))
	if err != nil {
		return nil, fmt.Errorf("ringpop creation failed: %v", err)
	}

	discoveryProvider, err := newDiscoveryProvider(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery provider %v", err)
	}

	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}
	rpw := NewRingpopWraper(rp, bootstrapOpts, logger)

	return NewRingpopMonitor(serviceName, rpw, logger), nil

}
