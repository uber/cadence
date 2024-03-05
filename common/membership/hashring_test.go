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

package membership

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomHostInfo(n int) []HostInfo {
	res := make([]HostInfo, n)
	for i := 0; i < n; i++ {
		res = append(res, NewHostInfo(randSeq(5)))
	}
	return res
}

func testCompareMembers(t *testing.T, curr []HostInfo, new []HostInfo, hasDiff bool) {
	hashring := &ring{}
	currMembers := make(map[string]HostInfo, len(curr))
	for _, m := range curr {
		currMembers[m.GetAddress()] = m
	}
	hashring.members.keys = currMembers
	newMembers, changed := hashring.compareMembers(new)
	assert.Equal(t, hasDiff, changed)
	assert.Equal(t, len(new), len(newMembers))
	for _, m := range new {
		_, ok := newMembers[m.GetAddress()]
		assert.True(t, ok)
	}
}

func Test_ring_compareMembers(t *testing.T) {

	tests := []struct {
		curr    []HostInfo
		new     []HostInfo
		hasDiff bool
	}{
		{curr: []HostInfo{}, new: []HostInfo{NewHostInfo("a")}, hasDiff: true},
		{curr: []HostInfo{}, new: []HostInfo{NewHostInfo("a"), NewHostInfo("b")}, hasDiff: true},
		{curr: []HostInfo{NewHostInfo("a")}, new: []HostInfo{NewHostInfo("a"), NewHostInfo("b")}, hasDiff: true},
		{curr: []HostInfo{}, new: []HostInfo{}, hasDiff: false},
		{curr: []HostInfo{NewHostInfo("a")}, new: []HostInfo{NewHostInfo("a")}, hasDiff: false},
		// order doesn't matter.
		{curr: []HostInfo{NewHostInfo("a"), NewHostInfo("b")}, new: []HostInfo{NewHostInfo("b"), NewHostInfo("a")}, hasDiff: false},
		// member has left the ring
		{curr: []HostInfo{NewHostInfo("a"), NewHostInfo("b"), NewHostInfo("c")}, new: []HostInfo{NewHostInfo("b"), NewHostInfo("a")}, hasDiff: true},
		// ring becomes empty
		{curr: []HostInfo{NewHostInfo("a"), NewHostInfo("b"), NewHostInfo("c")}, new: []HostInfo{}, hasDiff: true},
	}

	for _, tt := range tests {
		testCompareMembers(t, tt.curr, tt.new, tt.hasDiff)
	}

}

func TestFailedLookupWillAskProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	pp.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	pp.EXPECT().GetMembers("test-service").Times(1)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	hr.Start()
	_, err := hr.Lookup("a")

	assert.Error(t, err)
}

func TestRefreshUpdatesRingOnlyWhenRingHasChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	pp.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	pp.EXPECT().GetMembers("test-service").Times(3)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	hr.Start()

	hr.refresh()
	updatedAt := hr.members.refreshed
	hr.refresh()
	assert.Equal(t, updatedAt, hr.members.refreshed)

}

func TestSubscribeIgnoresDuplicates(t *testing.T) {
	var changeCh = make(chan *ChangedEvent)
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))

	assert.NoError(t, hr.Subscribe("test-service", changeCh))
	assert.Error(t, hr.Subscribe("test-service", changeCh))
	assert.Equal(t, 1, len(hr.subscribers.keys))
}

func TestUnsubcribeIgnoresDeletionOnEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	assert.Equal(t, 0, len(hr.subscribers.keys))
	assert.NoError(t, hr.Unsubscribe("test-service"))
	assert.NoError(t, hr.Unsubscribe("test-service"))
	assert.NoError(t, hr.Unsubscribe("test-service"))
}

func TestUnsubcribeDeletes(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)
	var changeCh = make(chan *ChangedEvent)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))

	assert.Equal(t, 0, len(hr.subscribers.keys))
	assert.NoError(t, hr.Subscribe("testservice1", changeCh))
	assert.Equal(t, 1, len(hr.subscribers.keys))
	assert.NoError(t, hr.Unsubscribe("test-service"))
	assert.Equal(t, 1, len(hr.subscribers.keys))
	assert.NoError(t, hr.Unsubscribe("testservice1"))
	assert.Equal(t, 0, len(hr.subscribers.keys))

}

func TestMemberCountReturnsNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	assert.Equal(t, 0, hr.MemberCount())

	ring := emptyHashring()
	for _, addr := range []string{"127", "128"} {
		host := NewHostInfo(addr)
		ring.AddMembers(host)
	}
	hr.value.Store(ring)
	assert.Equal(t, 2, hr.MemberCount())
}

func TestErrorIsPropagatedWhenProviderFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)
	pp.EXPECT().GetMembers(gomock.Any()).Return(nil, errors.New("error"))

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	assert.Error(t, hr.refresh())
}

func TestStopWillStopProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)

	pp.EXPECT().Stop().Times(1)

	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	hr.status = common.DaemonStatusStarted
	hr.Stop()

}

func TestLookupAndRefreshRaceCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	pp := NewMockPeerProvider(ctrl)
	var wg sync.WaitGroup

	pp.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	pp.EXPECT().GetMembers("test-service").AnyTimes().DoAndReturn(func(service string) ([]HostInfo, error) {
		return randomHostInfo(5), nil
	})
	hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))
	hr.Start()
	wg.Add(2)
	go func() {
		for i := 0; i < 50; i++ {
			_, _ = hr.Lookup("a")
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 50; i++ {
			// to bypass internal check
			hr.members.refreshed = time.Now().AddDate(0, 0, -1)
			assert.NoError(t, hr.refresh())
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestEmitHashringView(t *testing.T) {

	tests := map[string]struct {
		hosts          []HostInfo
		lookuperr      error
		selfInfo       HostInfo
		selfErr        error
		expectedResult float64
	}{
		"example one - sorted set 1 - the output should be some random hashed value": {
			hosts: []HostInfo{
				{addr: "10.0.0.1:1234", ip: "10.0.0.1", identity: "host1", portMap: nil},
				{addr: "10.0.0.2:1234", ip: "10.0.0.2", identity: "host2", portMap: nil},
				{addr: "10.0.0.3:1234", ip: "10.0.0.3", identity: "host3", portMap: nil},
			},
			selfInfo:       HostInfo{identity: "host123"},
			expectedResult: 835.0, // the number here is meaningless
		},
		"example one - unsorted set 1 - the order of the hosts should not matter": {
			hosts: []HostInfo{
				{addr: "10.0.0.1:1234", ip: "10.0.0.1", identity: "host1", portMap: nil},
				{addr: "10.0.0.3:1234", ip: "10.0.0.3", identity: "host3", portMap: nil},
				{addr: "10.0.0.2:1234", ip: "10.0.0.2", identity: "host2", portMap: nil},
			},
			selfInfo:       HostInfo{identity: "host123"},
			expectedResult: 835.0, // the test here is that it's the same as test 1
		},
		"example 2 - empty set": {
			hosts:          []HostInfo{},
			selfInfo:       HostInfo{identity: "host123"},
			expectedResult: 242.0, // meaningless hash value
		},
		"example 3 - nil set": {
			hosts:          nil,
			selfInfo:       HostInfo{identity: "host123"},
			expectedResult: 242.0, // meaningless hash value
		},
	}

	for name, td := range tests {

		t.Run(name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			pp := NewMockPeerProvider(ctrl)

			pp.EXPECT().GetMembers("test-service").DoAndReturn(func(service string) ([]HostInfo, error) {
				return td.hosts, td.lookuperr
			})

			pp.EXPECT().WhoAmI().DoAndReturn(func() (HostInfo, error) {
				return td.selfInfo, td.selfErr
			})

			hr := newHashring("test-service", pp, log.NewNoop(), metrics.NoopScope(0))

			assert.Equal(t, td.expectedResult, hr.emitHashIdentifier())
		})
	}
}
