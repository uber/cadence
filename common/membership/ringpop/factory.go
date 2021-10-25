// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ringpop

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/swim"
	"github.com/uber/tchannel-go"
	tchannel2 "go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/cadence/common/service"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/membership"
)

// Factory implements the Factory interface
type Factory struct {
	config      Config
	dispatcher  *yarpc.Dispatcher
	serviceName string
	logger      log.Logger

	sync.Mutex
	ringPop           *RingPop
	membershipMonitor membership.Monitor
}

func NewFactory(
	config Config,
	dispatcher *yarpc.Dispatcher,
	serviceName string,
	logger log.Logger,
) (*Factory, error) {

	if err := config.validate(); err != nil {
		return nil, err
	}
	if config.MaxJoinDuration == 0 {
		config.MaxJoinDuration = defaultMaxJoinDuration
	}
	return &Factory{
		config:      config,
		dispatcher:  dispatcher,
		serviceName: serviceName,
		logger:      logger,
	}, nil
}

// GetMembershipMonitor return a membership monitor
func (factory *Factory) GetMembershipMonitor() (membership.Monitor, error) {
	factory.Lock()
	defer factory.Unlock()

	return factory.getMembership()
}

func (factory *Factory) getMembership() (membership.Monitor, error) {
	if factory.membershipMonitor != nil {
		return factory.membershipMonitor, nil
	}

	membershipMonitor, err := factory.createMembership()
	if err != nil {
		return nil, err
	}
	factory.membershipMonitor = membershipMonitor
	return membershipMonitor, nil
}

func (factory *Factory) createMembership() (membership.Monitor, error) {
	// use actual listen port (in case service is bound to :0 or 0.0.0.0:0)
	rp, err := factory.getRingpop()
	if err != nil {
		return nil, fmt.Errorf("ringpop creation failed: %v", err)
	}

	return NewMonitor(factory.serviceName, service.List, rp, factory.logger), nil
}

func (factory *Factory) getRingpop() (*RingPop, error) {
	if factory.ringPop != nil {
		return factory.ringPop, nil
	}

	ringPop, err := factory.createRingpop()
	if err != nil {
		return nil, err
	}
	factory.ringPop = ringPop
	return ringPop, nil
}

func (factory *Factory) createRingpop() (*RingPop, error) {

	var ch *tchannel.Channel
	var err error
	if ch, err = factory.getChannel(factory.dispatcher); err != nil {
		return nil, err
	}

	rp, err := ringpop.New(factory.config.Name, ringpop.Channel(ch))
	if err != nil {
		return nil, err
	}

	discoveryProvider, err := newDiscoveryProvider(factory.config, factory.logger)
	if err != nil {
		return nil, err
	}
	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  factory.config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}
	return NewRingPop(rp, bootstrapOpts, factory.logger), nil
}

func (factory *Factory) getChannel(
	dispatcher *yarpc.Dispatcher,
) (*tchannel.Channel, error) {

	t := dispatcher.Inbounds()[0].Transports()[0].(*tchannel2.ChannelTransport)
	ty := reflect.ValueOf(t.Channel())
	var ch *tchannel.Channel
	var ok bool
	if ch, ok = ty.Interface().(*tchannel.Channel); !ok {
		return nil, errors.New("unable to get tchannel out of the dispatcher")
	}
	return ch, nil
}
