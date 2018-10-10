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

	"fmt"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

var historyTestRetryPolicy = createHistoryTestRetryPolicy()

func createHistoryTestRetryPolicy() backoff.RetryPolicy {
	policy := backoff.NewExponentialRetryPolicy(time.Millisecond * 50)
	policy.SetMaximumInterval(time.Second * 5)
	policy.SetExpirationInterval(time.Second * 15)

	return policy
}

func isConditionFail(err error) bool {
	switch err.(type) {
	case *p.ConditionFailedError:
		return true
	case *p.UnexpectedConditionFailedError:
		//TODO we need to understand why it can return UnexpectedConditionFailedError
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
	uuid, err := gocql.RandomUUID()
	s.Nil(err)
	return uuid.String()
}

// TestAppendHistoryEvents test
func (s *HistoryV2PersistenceSuite) TestNewBranch() {
	treeID := s.genRandomUUIDString()

	isNewCount := 0
	concurrency := 10
	for i := 0; i < concurrency; i++ {
		go func() {
			brID := s.genRandomUUIDString()
			IsNewTree, err := s.newHistoryBranch(treeID, brID)
			fmt.Println("err from newHistoryBranch", err)
			s.Nil(err)
			if IsNewTree {
				isNewCount++
			}
		}()
	}
	s.Equal(1, isNewCount)

	branches := s.descTree(treeID)
	s.Equal(concurrency, branches)

	for i := 0; i < concurrency; i++ {
		go func(brID string) {
			s.deleteHistoryBranch(treeID, brID)
		}(branches[i].BranchID)
	}

	branches = s.descTree(treeID)
	s.Equal(0, branches)

}

// NewHistoryBranch helper
func (s *HistoryV2PersistenceSuite) newHistoryBranch(treeID, branchID string) (bool, error) {

	var resp *p.NewHistoryBranchResponse

	op := func() error {
		var err error
		resp, err = s.HistoryMgr.NewHistoryBranch(&p.NewHistoryBranchRequest{
			BranchInfo: p.HistoryBranch{
				TreeID:   treeID,
				BranchID: branchID,
			},
		})
		return err
	}

	isNewT := false
	err := backoff.Retry(op, historyTestRetryPolicy, isConditionFail)
	if resp != nil {
		isNewT = resp.IsNewTree
	}
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
