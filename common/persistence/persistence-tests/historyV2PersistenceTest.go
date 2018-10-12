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

package persistencetests

import (
	"os"
	"testing"

	"time"

	"sync/atomic"

	"sync"

	"math/rand"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// HistoryV2PersistenceSuite contains history persistence tests
	HistoryV2PersistenceSuite struct {
		suite.Suite
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

var increasingVersion int64
var historyTestRetryPolicy = createHistoryTestRetryPolicy()

func createHistoryTestRetryPolicy() backoff.RetryPolicy {
	policy := backoff.NewExponentialRetryPolicy(time.Millisecond * 50)
	policy.SetMaximumInterval(time.Second * 3)
	policy.SetExpirationInterval(time.Second * 30)

	return policy
}

func isConditionFail(err error) bool {
	switch err.(type) {
	case *p.ConditionFailedError:
		return true
	default:
		return false
	}
}

// SetupSuite implementation
func (s *HistoryV2PersistenceSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// SetupTest implementation
func (s *HistoryV2PersistenceSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

// TearDownSuite implementation
func (s *HistoryV2PersistenceSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

func (s *HistoryV2PersistenceSuite) genRandomUUIDString() string {
	at := time.Unix(rand.Int63(), rand.Int63())
	uuid := gocql.UUIDFromTime(at)
	return uuid.String()
}

func (s *HistoryV2PersistenceSuite) TestConcurrentlyCreateAndDeleteEmptyBranches() {
	treeID := s.genRandomUUIDString()

	wg := sync.WaitGroup{}
	newCount := int32(0)
	concurrency := 10
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			brID := s.genRandomUUIDString()
			_, isNewTree, err := s.newHistoryBranch(treeID, brID)
			s.Nil(err)
			if isNewTree {
				atomic.AddInt32(&newCount, 1)
			}
		}()
	}

	wg.Wait()
	newCount = atomic.LoadInt32(&newCount)
	s.Equal(int32(1), newCount)

	branches := s.descTree(treeID)
	s.Equal(concurrency, len(branches))

	wg = sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(brID string) {
			defer wg.Done()
			// delete old branches along with create new branches
			err := s.deleteHistoryBranch(treeID, brID)
			s.Nil(err)

			brID2 := s.genRandomUUIDString()
			_, _, err = s.newHistoryBranch(treeID, brID2)
			s.Nil(err)
		}(branches[i].BranchID)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(10, len(branches))

	// delete the newly empty branches
	wg = sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(brID string) {
			defer wg.Done()
			err := s.deleteHistoryBranch(treeID, brID)
			s.Nil(err)
		}(branches[i].BranchID)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(0, len(branches))

	// create one more try after delete the whole tree
	brID := s.genRandomUUIDString()
	_, isNewTree, err := s.newHistoryBranch(treeID, brID)
	s.Nil(err)
	s.Equal(true, isNewTree)
	branches = s.descTree(treeID)
	s.Equal(1, len(branches))

	// a final clean up
	err = s.deleteHistoryBranch(treeID, brID)
	s.Nil(err)
	branches = s.descTree(treeID)
	s.Equal(0, len(branches))
}

func (s *HistoryV2PersistenceSuite) TestConcurrentlyCreateAndAppendBranches() {
	treeID := s.genRandomUUIDString()

	wg := sync.WaitGroup{}
	newCount := int32(0)
	concurrency := 10

	// test create new branch along with appending new nodes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			brID := s.genRandomUUIDString()
			bi, isNewTree, err := s.newHistoryBranch(treeID, brID)
			s.Nil(err)
			if isNewTree {
				atomic.AddInt32(&newCount, 1)
			}
			historyW := &workflow.History{}
			events := s.genRandomEvents([]int64{1, 2, 3})

			_, overrides, err := s.append(bi, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = events

			events = s.genRandomEvents([]int64{4})
			_, overrides, err = s.append(bi, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			//try empty ancestor
			bi = p.HistoryBranch{
				TreeID:   treeID,
				BranchID: brID,
			}

			events = s.genRandomEvents([]int64{5, 6, 7, 8})
			_, overrides, err = s.append(bi, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			events = s.genRandomEvents([]int64{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
			_, overrides, err = s.append(bi, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			//read branch to verify
			historyR := &workflow.History{}
			bi, events, token, err := s.read(bi, 1, 21, 0, 10, nil)
			s.Nil(err)
			s.Equal(10, len(events))
			historyR.Events = events

			_, events, token, err = s.read(bi, 1, 21, 0, 10, token)
			s.Nil(err)
			s.Equal(10, len(events))
			historyR.Events = append(historyR.Events, events...)

			// the next page should return empty events
			_, events, token, err = s.read(bi, 1, 21, 0, 10, token)
			s.Nil(err)
			s.Equal(0, len(events))

			s.True(historyW.Equals(historyR))
		}(i)
	}

	wg.Wait()
	newCount = atomic.LoadInt32(&newCount)
	s.Equal(int32(1), newCount)
	branches := s.descTree(treeID)
	s.Equal(concurrency, len(branches))

	// test appending nodes(override and new nodes) on each branch concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			brID := branches[idx].BranchID
			branch := p.HistoryBranch{
				TreeID:    treeID,
				BranchID:  brID,
				Ancestors: nil,
			}

			events := s.genRandomEvents([]int64{5})
			branch, overrides, err := s.append(branch, events, 1)
			s.Nil(err)
			s.Equal(1, overrides)
			// read to verify override
			_, events, _, err = s.read(branch, 1, 25, 0, 10, nil)
			s.Nil(err)
			s.Equal(5, len(events))

			events = s.genRandomEvents([]int64{6, 7, 8})
			branch, overrides, err = s.append(branch, events, 2)
			s.Nil(err)
			s.Equal(3, overrides)
			// read to verify override
			_, events, _, err = s.read(branch, 1, 25, 0, 10, nil)
			s.Nil(err)
			s.Equal(8, len(events))

			events = s.genRandomEvents([]int64{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23})
			branch, overrides, err = s.append(branch, events, 2)
			s.Nil(err)
			s.Equal(12, overrides)

			_, events, _, err = s.read(branch, 1, 25, 0, 25, nil)
			s.Nil(err)
			s.Equal(23, len(events))
		}(i)
	}

	wg.Wait()
	// Finally lets clean up all branches
	wg = sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(brID string) {
			defer wg.Done()
			err := s.deleteHistoryBranch(treeID, brID)
			s.Nil(err)
		}(branches[i].BranchID)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(0, len(branches))

}

func (s *HistoryV2PersistenceSuite) TestConcurrentlyForkAndAppendBranches() {
	treeID := s.genRandomUUIDString()

	wg := sync.WaitGroup{}
	concurrency := 10
	masterBrID := s.genRandomUUIDString()
	masterBr, isNewTree, err := s.newHistoryBranch(treeID, masterBrID)
	s.Nil(err)
	s.Equal(true, isNewTree)

	// append first batch to master branch
	eids := []int64{}
	for i := int64(1); i <= int64(concurrency); i++ {
		eids = append(eids, i)
	}
	events0 := s.genRandomEvents(eids)
	_, overrides, err := s.append(masterBr, events0, 0)
	s.Nil(err)
	s.Equal(0, overrides)

	level1 := sync.Map{}
	// test forking from master branch and append nodes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			brID := s.genRandomUUIDString()
			forkNodeID := rand.Int63n(int64(concurrency)) + 1
			level1.Store(brID, forkNodeID)

			bi, err := s.fork(masterBr, forkNodeID, brID)
			s.Nil(err)
			s.Equal(1, len(bi.Ancestors))
			s.Equal(masterBrID, bi.Ancestors[0].BranchID)

			// append second batch to first level
			eids = []int64{}
			for i := forkNodeID + 1; i <= int64(concurrency)*2; i++ {
				eids = append(eids, i)
			}
			events1 := s.genRandomEvents(eids)

			_, overrides, err := s.append(bi, events1, 0)
			s.Nil(err)
			s.Equal(0, overrides)

			_, events1, _, err = s.read(bi, 1, int64(concurrency)*2+1, 0, (concurrency)*2, nil)
			s.Nil(err)
			s.Equal((concurrency)*2, len(events1))
		}(i)
	}

	wg.Wait()
	branches := s.descTree(treeID)
	s.Equal(concurrency+1, len(branches))
	level1Brs := []p.HistoryBranch{}
	for _, bi := range branches {
		if bi.BranchID != masterBrID {
			level1Brs = append(level1Brs, bi)
		}
	}

	// test forking for second level of branch
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			brID := s.genRandomUUIDString()
			// so it is possible that the new branch will fork from master branch
			forkNodeID := rand.Int63n(int64(concurrency)*2) + 1
			forkBr := level1Brs[idx]
			v, ok := level1.Load(forkBr.BranchID)
			s.Equal(true, ok)
			lastForkNodeID, ok := v.(int64)
			s.Equal(true, ok)
			forkOnMaster := false
			if forkNodeID <= lastForkNodeID {
				forkOnMaster = true
			}

			bi, err := s.fork(forkBr, forkNodeID, brID)
			s.Nil(err)

			if forkOnMaster {
				s.Equal(1, len(bi.Ancestors))
				s.Equal(masterBrID, bi.Ancestors[0].BranchID)
			} else {
				s.Equal(2, len(bi.Ancestors))
			}

			// append second batch to second level
			eids = []int64{}
			for i := forkNodeID + 1; i <= int64(concurrency)*3; i++ {
				eids = append(eids, i)
			}
			events2 := s.genRandomEvents(eids)

			_, overrides, err := s.append(bi, events2, 0)
			s.Nil(err)
			s.Equal(0, overrides)

			_, events2, _, err = s.read(bi, 1, int64(concurrency)*3+1, 0, (concurrency)*3, nil)
			s.Nil(err)
			s.Equal((concurrency)*3, len(events2))

			//test fork and newBranch concurrently
			brID2 := s.genRandomUUIDString()
			_, _, err = s.newHistoryBranch(treeID, brID2)
			s.Nil(err)
		}(i)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(int(concurrency*3+1), len(branches))
	// Finally lets clean up all branches
	wg = sync.WaitGroup{}
	for i := 0; i < concurrency*3+1; i++ {
		wg.Add(1)
		go func(brID string) {
			defer wg.Done()
			err := s.deleteHistoryBranch(treeID, brID)
			s.Nil(err)
		}(branches[i].BranchID)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(0, len(branches))

}

func (s *HistoryV2PersistenceSuite) genRandomEvents(eventIDs []int64) []*workflow.HistoryEvent {
	atomic.AddInt64(&increasingVersion, 1)
	var events []*workflow.HistoryEvent

	timestamp := time.Now().UnixNano()
	for _, eid := range eventIDs {
		e := &workflow.HistoryEvent{EventId: common.Int64Ptr(eid), Version: common.Int64Ptr(increasingVersion), Timestamp: int64Ptr(timestamp)}
		events = append(events, e)
	}

	return events
}

// persistence helper
func (s *HistoryV2PersistenceSuite) newHistoryBranch(treeID, branchID string) (p.HistoryBranch, bool, error) {

	isNewT := false
	var bi p.HistoryBranch

	op := func() error {
		var err error
		resp, err := s.HistoryMgr.NewHistoryBranch(&p.NewHistoryBranchRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
		})
		if resp != nil && resp.IsNewTree {
			isNewT = true
		}
		if resp != nil {
			bi = resp.BranchInfo
		}
		return err
	}

	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	return bi, isNewT, err
}

// persistence helper
func (s *HistoryV2PersistenceSuite) deleteHistoryBranch(treeID, branchID string) error {

	op := func() error {
		var err error
		err = s.HistoryMgr.DeleteHistoryBranch(&p.DeleteHistoryBranchRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
		})
		return err
	}

	return backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
}

// persistence helper
func (s *HistoryV2PersistenceSuite) descTree(treeID string) []p.HistoryBranch {
	resp, err := s.HistoryMgr.GetHistoryTree(&p.GetHistoryTreeRequest{
		TreeID: treeID,
	})
	s.Nil(err)
	return resp.Branches
}

// persistence helper
func (s *HistoryV2PersistenceSuite) read(branch p.HistoryBranch, minID, maxID, lastVersion int64, pageSize int, token []byte) (p.HistoryBranch, []*workflow.HistoryEvent, []byte, error) {

	resp, err := s.HistoryMgr.ReadHistoryBranch(&p.ReadHistoryBranchRequest{
		BranchInfo:       branch,
		MinEventID:       minID,
		MaxEventID:       maxID,
		PageSize:         pageSize,
		NextPageToken:    token,
		LastEventVersion: lastVersion,
	})
	if err != nil {
		return p.HistoryBranch{}, nil, nil, err
	}
	if len(resp.History) > 0 {
		s.True(resp.Size > 0)
	}
	return resp.BranchInfo, resp.History, resp.NextPageToken, nil
}

// persistence helper
func (s *HistoryV2PersistenceSuite) append(bi p.HistoryBranch, events []*workflow.HistoryEvent, txnID int64) (p.HistoryBranch, int, error) {

	var resp *p.AppendHistoryNodesResponse

	op := func() error {
		var err error
		resp, err = s.HistoryMgr.AppendHistoryNodes(&p.AppendHistoryNodesRequest{
			BranchInfo:    bi,
			Events:        events,
			TransactionID: txnID,
			Encoding:      pickRandomEncoding(),
		})
		return err
	}

	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	if err != nil {
		return p.HistoryBranch{}, 0, err
	}
	s.True(resp.Size > 0)

	return resp.BranchInfo, resp.OverrideCount, err
}

// persistence helper
func (s *HistoryV2PersistenceSuite) fork(forkBranch p.HistoryBranch, forkNodeID int64, newBranchID string) (p.HistoryBranch, error) {

	bi := p.HistoryBranch{}

	op := func() error {
		var err error
		resp, err := s.HistoryMgr.ForkHistoryBranch(&p.ForkHistoryBranchRequest{
			BranchInfo:     forkBranch,
			ForkFromNodeID: forkNodeID,
			NewBranchID:    newBranchID,
		})
		if resp != nil {
			bi = resp.BranchInfo
		}
		return err
	}

	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	return bi, err
}
