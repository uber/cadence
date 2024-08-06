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
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/service/matching/config"
	"github.com/uber/cadence/service/matching/handler"
)

func TestMembershipSubscriptionShutdown(t *testing.T) {
	assert.NotPanics(t, func() {
		ctrl := gomock.NewController(t)
		// this is nil for the memebership resolver, and it'll panic
		r := resource.NewTest(t, ctrl, 0)
		s := Service{
			stopC:    make(chan struct{}),
			Resource: r,
		}
		e := handler.NewMockEngine(ctrl)
		r.MembershipResolver.EXPECT().Subscribe(service.Matching, "matching-engine", gomock.Any()).Times(1)
		r.MembershipResolver.EXPECT().WhoAmI().Times(1)

		go func() {
			close(s.stopC)
		}()
		s.subscribeToMembershipChanges(e)
	})
}

func TestMembershipSubscriptionPanicHandling(t *testing.T) {
	assert.NotPanics(t, func() {
		ctrl := gomock.NewController(t)
		// this is nil for the memebership resolver, and it'll panic
		r := resource.NewTest(t, ctrl, 0)
		s := Service{
			stopC:    make(chan struct{}),
			Resource: r,
		}
		e := handler.NewMockEngine(ctrl)
		r.MembershipResolver.EXPECT().Subscribe(service.Matching, "matching-engine", gomock.Any()).DoAndReturn(func(_, _, _ any) {
			panic("a panic has occurred")
		})
		s.subscribeToMembershipChanges(e)
	})
}

func TestCheckIfMembershipStopped(t *testing.T) {

	m := &mockEng{}
	hostinfo := membership.NewDetailedHostInfo("", "some-host-A", nil)
	c := &config.Config{}

	c.EnableServiceDiscoveryShutdown = func(opts ...dynamicconfig.FilterOption) bool { return true }

	changeEvent := membership.ChangedEvent{
		HostsRemoved: []string{"some-host-A"},
	}

	checkAndShutdownIfEvicted(m, c, log.NewNoop(), hostinfo, &changeEvent)
	assert.True(t, m.stopped)
	assert.False(t, m.started)
}

func TestCheckIfMembershipRestarted(t *testing.T) {

	m := &mockEng{}
	hostinfo := membership.NewDetailedHostInfo("", "some-host-A", nil)
	c := &config.Config{}

	c.EnableServiceDiscoveryShutdown = func(opts ...dynamicconfig.FilterOption) bool { return true }

	changeEvent := membership.ChangedEvent{
		HostsAdded: []string{"some-host-A"},
	}

	checkAndShutdownIfEvicted(m, c, log.NewNoop(), hostinfo, &changeEvent)
	assert.False(t, m.stopped)
	assert.True(t, m.started)
}

type mockEng struct {
	stopped bool
	started bool
}

func (m *mockEng) Stop() {
	m.stopped = true
}

func (m *mockEng) Start() {
	m.started = true
}
