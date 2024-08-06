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

package matching

import (
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/service/matching/config"
	"github.com/uber/cadence/service/matching/handler"
)

func (s *Service) subscribeToMembershipChanges(
	engine handler.Engine,
) {
	defer func() {
		if r := recover(); r != nil {
			s.GetLogger().Error("matching membership watcher changes caused a panic, recovering", tag.Dynamic("recovered-panic", r))
		}
	}()
	listener := make(chan *membership.ChangedEvent)
	s.GetMembershipResolver().Subscribe(service.Matching, "matching-engine", listener)
	self, err := s.GetMembershipResolver().WhoAmI()
	if err != nil {
		s.GetLogger().Error("Failed to get self from hashring")
		return
	}
	for {
		select {
		case event := <-listener:
			checkAndShutdownIfEvicted(engine, s.config, s.GetLogger(), self, event)
		case <-s.stopC:
			return
		}
	}
}

// checkAndShutdownIfEvicted checks if the host has been evicted from the hashring
func checkAndShutdownIfEvicted(engine common.Daemon, cfg *config.Config, log log.Logger, self membership.HostInfo, event *membership.ChangedEvent) {
	featureEnabled := cfg.EnableServiceDiscoveryShutdown()
	if selfIsEvicted(log, self, event) {
		if featureEnabled {
			log.Info("shutting down service due to hashring removing it from service", tag.MembershipChangeEvent(event))
			engine.Stop()
		} else {
			log.Info("Service has been removed from hashring. No other actions triggered", tag.MembershipChangeEvent(event))
		}
	}
	// these are hopefully mutually exclusive,
	if selfIsReturned(log, self, event) {
		if featureEnabled {
			log.Info("service has been returned to hashring and engine restarted", tag.MembershipChangeEvent(event))
			engine.Start()
		} else {
			log.Info("Service has returned to the hashring. No other actions triggered", tag.MembershipChangeEvent(event))
		}
	}
}

// This is the expected case where the service discovery might race and remove the instance
// during the host shutdown.
// this looks at the changes of hosts, eg:
// { Removals["10.90.37.96_31773_31771"] }
// and reconstructs the to ports and check if they match self
func selfIsEvicted(log log.Logger, self membership.HostInfo, updates *membership.ChangedEvent) bool {
	for _, r := range updates.HostsRemoved {
		matches := r == self.Identity()
		log.Debug("Container removal check: determining if this host needs to be removed from the ring and/or shutdown",
			tag.MembershipChangeEvent(updates),
			tag.Dynamic("self-hostport", self.String()),
			tag.Dynamic("removal-result", matches),
		)
		if matches {
			return true
		}
	}
	return false
}

// this is somewhat an unexpected case and is added here as a best effort.
// by the time the subscriptino is called, it's expected that it'll probably be healthy and live
// but for symmetry and cases where there might be drains or flickr, allow the engine to be reenabled
func selfIsReturned(log log.Logger, self membership.HostInfo, updates *membership.ChangedEvent) bool {
	for _, r := range updates.HostsAdded {
		matches := r == self.Identity()
		log.Debug("Container return check: determining if this host needs to be removed from the ring and/or shutdown",
			tag.MembershipChangeEvent(updates),
			tag.Dynamic("self-hostport", self.String()),
			tag.Dynamic("removal-result", matches),
		)
		if matches {
			return true
		}
	}
	return false
}
