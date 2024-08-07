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

package handler

import (
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
)

// Because there's a bunch of conditions under which matching may be holding a tasklist
// reader daemon and other live procesess but when it doesn't (according to the rest of the hashring)
// own the tasklist anymore, this listener watches for membership changes and purges anything disused
// in the hashring on membership changes.
//
// Combinedd with the guard on tasklist instantiation, it should prevent incorrect or poorly timed
// creating of tasklist ownership and database shard thrashing between hosts while they figure out
// which host is the real owner of the tasklist.
//
// This is not the main shutdown process, its just an optimization.
func (e *matchingEngineImpl) subscribeToMembershipChanges() {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("matching membership watcher changes caused a panic, recovering", tag.Dynamic("recovered-panic", r))
		}
	}()

	if !e.config.EnableTasklistOwnershipGuard() {
		return
	}

	listener := make(chan *membership.ChangedEvent, 1000)
	e.membershipResolver.Subscribe(service.Matching, "matching-engine", listener)

	for {
		select {
		case event := <-listener:
			err := e.shutDownNonOwnedTasklists()
			if err != nil {
				e.logger.Error("Error while trying to determine if tasklists have been shutdown",
					tag.Error(err),
					tag.MembershipChangeEvent(event),
				)
			}
		case <-e.shutdown:
			return
		}
	}
}
