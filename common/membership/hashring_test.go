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
	"go.uber.org/goleak"
	"go.uber.org/zap/zaptest/observer"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log/testlogger"
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
	res := make([]HostInfo, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, NewDetailedHostInfo(randSeq(5), randSeq(12), PortMap{randSeq(3): 666}))
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

type hashringTestData struct {
	t                *testing.T
	mockPeerProvider *MockPeerProvider
	mockTimeSource   clock.MockedTimeSource
	hashRing         *ring
	observedLogs     *observer.ObservedLogs
}

func newHashringTestData(t *testing.T) *hashringTestData {
	var td hashringTestData

	ctrl := gomock.NewController(t)
	td.t = t
	td.mockPeerProvider = NewMockPeerProvider(ctrl)
	td.mockTimeSource = clock.NewMockedTimeSourceAt(time.Now())

	logger, observedLogs := testlogger.NewObserved(t)
	td.observedLogs = observedLogs

	td.hashRing = newHashring(
		"test-service",
		td.mockPeerProvider,
		td.mockTimeSource,
		logger,
		metrics.NoopScope(0),
	)

	return &td
}

// starts hashring' background work and verifies all the goroutines closed at the end
func (td *hashringTestData) startHashRing() {
	td.mockPeerProvider.EXPECT().Stop()

	td.t.Cleanup(func() {
		td.hashRing.Stop()
		goleak.VerifyNone(td.t)
	})

	td.hashRing.Start()
}

func TestFailedLookupWillAskProvider(t *testing.T) {
	td := newHashringTestData(t)

	td.mockPeerProvider.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	td.mockPeerProvider.EXPECT().GetMembers("test-service").Times(1)

	td.startHashRing()
	_, err := td.hashRing.Lookup("a")

	assert.Error(t, err)
}

func TestRefreshUpdatesRingOnlyWhenRingHasChanged(t *testing.T) {
	td := newHashringTestData(t)

	td.mockPeerProvider.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	td.mockPeerProvider.EXPECT().GetMembers("test-service").Times(1).Return(randomHostInfo(3), nil)

	// Start will also call .refresh()
	td.startHashRing()
	updatedAt := td.hashRing.members.refreshed
	td.hashRing.refresh()
	refreshed, err := td.hashRing.refresh()

	assert.NoError(t, err)
	assert.False(t, refreshed)
	assert.Equal(t, updatedAt, td.hashRing.members.refreshed)
}

func TestRefreshWillNotifySubscribers(t *testing.T) {
	td := newHashringTestData(t)

	var hostsToReturn []HostInfo
	td.mockPeerProvider.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	td.mockPeerProvider.EXPECT().GetMembers("test-service").Times(2).DoAndReturn(func(service string) ([]HostInfo, error) {
		hostsToReturn = randomHostInfo(5)
		return hostsToReturn, nil
	})

	changed := &ChangedEvent{
		HostsAdded:   []string{"a"},
		HostsUpdated: []string{"b"},
		HostsRemoved: []string{"c"},
	}

	td.startHashRing()

	var changeCh = make(chan *ChangedEvent, 2)
	// Check if multiple subscribers will get notified
	assert.NoError(t, td.hashRing.Subscribe("subscriber1", changeCh))
	assert.NoError(t, td.hashRing.Subscribe("subscriber2", changeCh))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		changedEvent := <-changeCh
		changedEvent2 := <-changeCh
		assert.Equal(t, changed, changedEvent)
		assert.Equal(t, changed, changedEvent2)
	}()

	// to bypass internal check
	td.hashRing.members.refreshed = time.Now().AddDate(0, 0, -1)
	td.hashRing.refreshChan <- changed
	wg.Wait() // wait until both subscribers will get notification
	// Test if internal members are updated
	assert.ElementsMatch(t, td.hashRing.Members(), hostsToReturn, "members should contain just-added nodes")
}

func TestSubscribersAreNotifiedPeriodically(t *testing.T) {
	td := newHashringTestData(t)

	var hostsToReturn []HostInfo

	td.mockPeerProvider.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	td.mockPeerProvider.EXPECT().GetMembers("test-service").Times(3).DoAndReturn(func(service string) ([]HostInfo, error) {
		// we have to change members since subscribers are only notified on change
		hostsToReturn = randomHostInfo(5)
		return hostsToReturn, nil
	})
	td.mockPeerProvider.EXPECT().WhoAmI().AnyTimes()

	td.startHashRing()

	var changeCh = make(chan *ChangedEvent, 1)
	assert.NoError(t, td.hashRing.Subscribe("subscriber1", changeCh))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		event := <-changeCh
		assert.Empty(t, event, "event should be empty when periodical update happens")
	}()

	td.mockTimeSource.BlockUntil(1)                   // we should wait until ticker(defaultRefreshInterval) is created
	td.mockTimeSource.Advance(defaultRefreshInterval) // and only then to advance time

	wg.Wait() // wait until subscriber will get notification

	// Test if internal members are updated
	assert.ElementsMatch(t, td.hashRing.Members(), hostsToReturn, "members should contain just-added nodes")
}

func TestSubscribeIgnoresDuplicates(t *testing.T) {
	var changeCh = make(chan *ChangedEvent)
	td := newHashringTestData(t)

	assert.NoError(t, td.hashRing.Subscribe("test-service", changeCh))
	assert.Error(t, td.hashRing.Subscribe("test-service", changeCh))
	assert.Equal(t, 1, len(td.hashRing.subscribers.keys))
}

func TestUnsubcribeIgnoresDeletionOnEmpty(t *testing.T) {
	td := newHashringTestData(t)

	assert.Equal(t, 0, len(td.hashRing.subscribers.keys))
	assert.NoError(t, td.hashRing.Unsubscribe("test-service"))
	assert.NoError(t, td.hashRing.Unsubscribe("test-service"))
	assert.NoError(t, td.hashRing.Unsubscribe("test-service"))
}

func TestUnsubcribeDeletes(t *testing.T) {
	td := newHashringTestData(t)
	var changeCh = make(chan *ChangedEvent)

	assert.Equal(t, 0, len(td.hashRing.subscribers.keys))
	assert.NoError(t, td.hashRing.Subscribe("testservice1", changeCh))
	assert.Equal(t, 1, len(td.hashRing.subscribers.keys))
	assert.NoError(t, td.hashRing.Unsubscribe("test-service"))
	assert.Equal(t, 1, len(td.hashRing.subscribers.keys))
	assert.NoError(t, td.hashRing.Unsubscribe("testservice1"))
	assert.Equal(t, 0, len(td.hashRing.subscribers.keys))

}

func TestMemberCountReturnsNumber(t *testing.T) {
	td := newHashringTestData(t)

	assert.Equal(t, 0, td.hashRing.MemberCount())

	ring := emptyHashring()
	for _, addr := range []string{"127", "128"} {
		host := NewHostInfo(addr)
		ring.AddMembers(host)
	}
	td.hashRing.value.Store(ring)
	assert.Equal(t, 2, td.hashRing.MemberCount())
}

func TestErrorIsPropagatedWhenProviderFails(t *testing.T) {
	td := newHashringTestData(t)

	td.mockPeerProvider.EXPECT().GetMembers(gomock.Any()).Return(nil, errors.New("error"))

	_, err := td.hashRing.refresh()
	assert.Error(t, err)
}

func TestStopWillStopProvider(t *testing.T) {
	td := newHashringTestData(t)

	td.mockPeerProvider.EXPECT().Stop().Times(1)

	td.hashRing.status = common.DaemonStatusStarted
	td.hashRing.Stop()
}

func TestLookupAndRefreshRaceCondition(t *testing.T) {
	td := newHashringTestData(t)
	var wg sync.WaitGroup

	td.mockPeerProvider.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Times(1)
	td.mockPeerProvider.EXPECT().GetMembers("test-service").AnyTimes().DoAndReturn(func(service string) ([]HostInfo, error) {
		return randomHostInfo(5), nil
	})

	td.startHashRing()
	wg.Add(2)
	go func() {
		for i := 0; i < 50; i++ {
			_, _ = td.hashRing.Lookup("a")
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 50; i++ {
			// to bypass internal check
			td.hashRing.members.refreshed = time.Now().AddDate(0, 0, -1)
			_, err := td.hashRing.refresh()
			assert.NoError(t, err)
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

	for testName, testInput := range tests {
		t.Run(testName, func(t *testing.T) {
			td := newHashringTestData(t)

			td.mockPeerProvider.EXPECT().GetMembers("test-service").DoAndReturn(func(service string) ([]HostInfo, error) {
				return testInput.hosts, testInput.lookuperr
			})

			td.mockPeerProvider.EXPECT().WhoAmI().DoAndReturn(func() (HostInfo, error) {
				return testInput.selfInfo, testInput.selfErr
			})
			assert.Equal(t, testInput.expectedResult, td.hashRing.emitHashIdentifier())
		})
	}
}
