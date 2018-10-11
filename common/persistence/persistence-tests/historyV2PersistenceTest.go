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
			isNewTree, err := s.newHistoryBranch(treeID, brID)
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
			err := s.deleteHistoryBranch(treeID, brID)
			s.Nil(err)
		}(branches[i].BranchID)
	}

	wg.Wait()
	branches = s.descTree(treeID)
	s.Equal(0, len(branches))

	// create one more try after delete the whole tree
	brID := s.genRandomUUIDString()
	isNewTree, err := s.newHistoryBranch(treeID, brID)
	s.Nil(err)
	s.Equal(true, isNewTree)
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
			isNewTree, err := s.newHistoryBranch(treeID, brID)
			s.Nil(err)
			if isNewTree {
				atomic.AddInt32(&newCount, 1)
			}
			historyW := &workflow.History{}
			events := s.genRandomEvents([]int64{1, 2, 3})

			overrides, err := s.append(treeID, brID, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = events

			events = s.genRandomEvents([]int64{4})
			overrides, err = s.append(treeID, brID, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			events = s.genRandomEvents([]int64{5, 6, 7, 8})
			overrides, err = s.append(treeID, brID, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			events = s.genRandomEvents([]int64{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
			overrides, err = s.append(treeID, brID, events, 0)
			s.Nil(err)
			s.Equal(0, overrides)
			historyW.Events = append(historyW.Events, events...)

			//read branch to verify
			branch := p.HistoryBranch{
				TreeID:    treeID,
				BranchID:  brID,
				Ancestors: nil,
			}
			historyR := &workflow.History{}
			bi, events, token, err := s.read(branch, 1, 21, 0, 10, nil)
			s.Nil(err)
			s.Equal(10, len(events))
			historyR.Events = events

			_, events, token, err = s.read(*bi, 1, 21, 0, 10, token)
			s.Nil(err)
			s.Equal(10, len(events))
			historyR.Events = append(historyR.Events, events...)

			// the next page should return empty events
			_, events, token, err = s.read(*bi, 1, 21, 0, 10, token)
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
			overrides, err := s.append(treeID, brID, events, 1)
			s.Nil(err)
			s.Equal(1, overrides)
			// read to verify override
			_, events, _, err = s.read(branch, 1, 25, 0, 10, nil)
			s.Nil(err)
			s.Equal(5, len(events))

			events = s.genRandomEvents([]int64{6, 7, 8})
			overrides, err = s.append(treeID, brID, events, 2)
			s.Nil(err)
			s.Equal(3, overrides)
			// read to verify override
			_, events, _, err = s.read(branch, 1, 25, 0, 10, nil)
			s.Nil(err)
			s.Equal(8, len(events))

			events = s.genRandomEvents([]int64{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23})
			overrides, err = s.append(treeID, brID, events, 2)
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

// NewHistoryBranch helper
func (s *HistoryV2PersistenceSuite) newHistoryBranch(treeID, branchID string) (bool, error) {

	isNewT := false

	op := func() error {
		var err error
		resp, err := s.HistoryMgr.NewHistoryBranch(&p.NewHistoryBranchRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
		})
		//fmt.Println("newBr ret:", resp, err, treeID, branchID)
		if resp != nil && resp.IsNewTree {
			isNewT = true
		}
		return err
	}

	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	return isNewT, err
}

func (s *HistoryV2PersistenceSuite) deleteHistoryBranch(treeID, branchID string) error {

	op := func() error {
		var err error
		err = s.HistoryMgr.DeleteHistoryBranch(&p.DeleteHistoryBranchRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
		})
		//fmt.Println("delete ret:", err, treeID, branchID)
		return err
	}

	return backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
}

func (s *HistoryV2PersistenceSuite) descTree(treeID string) []p.HistoryBranch {
	resp, err := s.HistoryMgr.GetHistoryTree(&p.GetHistoryTreeRequest{
		TreeID: treeID,
	})
	s.Nil(err)
	return resp.Branches
}

func (s *HistoryV2PersistenceSuite) read(branch p.HistoryBranch, minID, maxID, lastVersion int64, pageSize int, token []byte) (*p.HistoryBranch, []*workflow.HistoryEvent, []byte, error) {

	resp, err := s.HistoryMgr.ReadHistoryBranch(&p.ReadHistoryBranchRequest{
		BranchInfo:       branch,
		MinEventID:       minID,
		MaxEventID:       maxID,
		PageSize:         pageSize,
		NextPageToken:    token,
		LastEventVersion: lastVersion,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if len(resp.History) > 0 {
		s.True(resp.Size > 0)
	}
	return &resp.BranchInfo, resp.History, resp.NextPageToken, nil
}

func (s *HistoryV2PersistenceSuite) append(treeID, branchID string, events []*workflow.HistoryEvent, txnID int64) (int, error) {

	var resp *p.AppendHistoryNodesResponse

	op := func() error {
		var err error
		resp, err = s.HistoryMgr.AppendHistoryNodes(&p.AppendHistoryNodesRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
			Events:        events,
			TransactionID: txnID,
			Encoding:      pickRandomEncoding(),
		})
		return err
	}

	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	if err != nil {
		return 0, err
	}
	s.True(resp.Size > 0)

	return resp.OverrideCount, err
}
