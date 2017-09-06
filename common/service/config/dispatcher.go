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

package config

import (
	"fmt"
	"net"

	"github.com/uber-common/bark"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
)

// DispatcherFactory is an implementation of service.DispatcherFactory interface
type DispatcherFactory struct {
	config      *RPC
	serviceName string
	ch          *tchannel.ChannelTransport
	logger      bark.Logger
}

// NewFactory builds a new DispatcherFactory
// conforming to the underlying configuration
func (cfg *RPC) NewFactory(sName string, logger bark.Logger) *DispatcherFactory {
	return newDispatcherFactory(cfg, sName, logger)
}

func newDispatcherFactory(cfg *RPC, sName string, logger bark.Logger) *DispatcherFactory {
	factory := &DispatcherFactory{config: cfg, serviceName: sName, logger: logger}
	return factory
}

// CreateDispatcher creates a dispatcher for inbound
func (d *DispatcherFactory) CreateDispatcher() *yarpc.Dispatcher {
	// Setup dispatcher for onebox
	var err error
	hostAddress := fmt.Sprintf("%v:%v", d.getListenIP(), d.config.Port)
	d.ch, err = tchannel.NewChannelTransport(
		tchannel.ServiceName(d.serviceName),
		tchannel.ListenAddr(hostAddress))
	if err != nil {
		d.logger.WithField("error", err).Fatal("Failed to create transport channel")
	}
	d.logger.Infof("Created YARPC dispatcher for: %v and listening at: %v",
		d.serviceName, hostAddress)
	return yarpc.NewDispatcher(yarpc.Config{
		Name:     d.serviceName,
		Inbounds: yarpc.Inbounds{d.ch.NewInbound()},
	})
}

// CreateDispatcherForOutbound creates a dispatcher for outbound connection
func (d *DispatcherFactory) CreateDispatcherForOutbound(
	callerName, serviceName, hostName string) *yarpc.Dispatcher {
	// Setup dispatcher(outbound) for onebox
	d.logger.Infof("Created YARPC dispatcher outbound for service: %v for host: %v",
		serviceName, hostName)
	return yarpc.NewDispatcher(yarpc.Config{
		Name: callerName,
		Outbounds: yarpc.Outbounds{
			serviceName: {Unary: d.ch.NewSingleOutbound(hostName)},
		},
	})
}

func (d *DispatcherFactory) getListenIP() net.IP {
	if d.config.BindOnLocalHost {
		return net.IPv4(127, 0, 0, 1)
	}
	ip, err := ListenIP()
	if err != nil {
		d.logger.Fatalf("tchannel.ListenIP failed, err=%v", err)
	}
	return ip
}
