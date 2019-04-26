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

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/membership"
	"go.uber.org/yarpc"
)

// SerfFactory implements the SerfFactory interface
type SerfFactory struct {
	config      *Serf
	logger      log.Logger
	serviceName string
}

// NewFactory builds a ringpop factory conforming
// to the underlying configuration
func (serfConfig *Serf) NewFactory(logger log.Logger, serviceName string) (*SerfFactory, error) {
	return &SerfFactory{serfConfig, logger, serviceName}, nil
}

// Create - creates a membership monitor backed by serf
func (s *SerfFactory) Create(dispatcher *yarpc.Dispatcher) (membership.Monitor, error) {
	config := serf.DefaultConfig()

	// set the hostname as identifier for serf
	config.NodeName = s.config.Name
	config.Tags = map[string]string{membership.RoleKey: s.serviceName}
	config.MemberlistConfig = memberlist.DefaultLANConfig()
	config.MemberlistConfig.BindAddr = "127.0.0.1"
	config.MemberlistConfig.BindPort = int(s.config.Port)
	config.MemberlistConfig.AdvertiseAddr = "127.0.0.1"
	config.MemberlistConfig.AdvertisePort = int(s.config.Port)
	s.logger.Info(fmt.Sprintf("creating serf monitor on port %v for service %v", s.config.Port, s.serviceName))
	return membership.NewSerfMonitor(CadenceServices, s.config.BootstrapHosts, config, s.logger), nil
}
