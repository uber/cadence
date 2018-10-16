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

	"math/rand"
	"time"

	"fmt"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// HistoryV2PersistenceSuite contains history persistence tests
	HistoryPerfSuite struct {
		suite.Suite
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

// SetupSuite implementation
func (s *HistoryPerfSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// SetupTest implementation
func (s *HistoryPerfSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

// TearDownSuite implementation
func (s *HistoryPerfSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

func (s *HistoryPerfSuite) genRandomUUIDString() string {
	at := time.Unix(rand.Int63(), rand.Int63())
	uuid := gocql.UUIDFromTime(at)
	return uuid.String()
}

func (s *HistoryPerfSuite) startProfile() int64 {
	return time.Now().UnixNano()
}

func (s *HistoryPerfSuite) stopProfile(startT int64, name string) {
	du := time.Now().UnixNano() - startT
	fmt.Printf("%v time eslapsed: %v milliseconds\n", name, float64(du)/float64(1000000))
}

func (s *HistoryPerfSuite) TestPerf() {
	treeID := s.genRandomUUIDString()
	brID := s.genRandomUUIDString()

	br, err := s.newHistoryBranch(treeID, brID)
	s.Nil(err)
	s.NotNil(br)

	//for v1
	domainID := treeID
	wf := workflow.WorkflowExecution{
		WorkflowId: &brID,
		RunId:      &brID,
	}

	total := 1000
	sizes := []int{1, 5, 10, 100, 1000}
	//1. test append 4*1000 events:
	//batch size = 1
	//batch size = 5
	//batch size = 10
	//batch size = 100
	//batch size = 1000
	firstIDV1 := int64(1)
	firstIDV2 := int64(1)

	for _, size := range sizes {

		st := s.startProfile()
		for i := 0; i < total/size; i++ {
			lastID := firstIDV1 + int64(size)
			events := s.genRandomEvents(firstIDV1, lastID)
			history := &workflow.History{
				Events: events,
			}

			err := s.appendV1(domainID, wf, firstIDV1, 0, 0, 0, history, false)
			s.Nil(err)

			firstIDV1 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("appendV1-batch size: %v", size))

		st = s.startProfile()
		for i := 0; i < total/size; i++ {
			lastID := firstIDV2 + int64(size)
			events := s.genRandomEvents(firstIDV2, lastID)

			_, ow, err := s.appendV2(br, events, 0)

			s.Equal(ow, 0)
			s.Nil(err)
			firstIDV2 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("appendV2-batch size: %v", size))

	}

	firstIDV1 = int64(1)
	firstIDV2 = int64(1)
	//2. test read 4*1000 events:
	for _, size := range sizes {

		st := s.startProfile()
		for i := 0; i < total/size; i++ {
			lastID := firstIDV1 + int64(size)

			historyR, _, err := s.readV1(domainID, wf, firstIDV1, lastID, int(lastID-firstIDV1), nil)
			s.Equal(size, len(historyR.Events))
			s.Nil(err)

			firstIDV1 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("readV1-batch size: %v", size))

		st = s.startProfile()
		for i := 0; i < total/size; i++ {
			lastID := firstIDV2 + int64(size)

			_, events, _, err := s.readv2(br, firstIDV2, lastID, 0, int(lastID-firstIDV2), nil)

			s.Equal(size, len(events))
			s.Nil(err)
			firstIDV2 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("readv2-batch size: %v", size))

	}
}

// lastID is exclusive
func (s *HistoryPerfSuite) genRandomEvents(firstID, lastID int64) []*workflow.HistoryEvent {
	events := make([]*workflow.HistoryEvent, 0, lastID-firstID+1)

	timestamp := time.Now().UnixNano()
	for eid := firstID; eid < lastID; eid++ {
		e := &workflow.HistoryEvent{EventId: common.Int64Ptr(eid), Version: common.Int64Ptr(timestamp), Timestamp: int64Ptr(timestamp)}
		events = append(events, e)
	}

	return events
}

// persistence helper
func (s *HistoryPerfSuite) newHistoryBranch(treeID, branchID string) (p.HistoryBranch, error) {

	resp, err := s.HistoryMgr.NewHistoryBranch(&p.NewHistoryBranchRequest{
		BranchInfo: p.HistoryBranch{
			TreeID:   treeID,
			BranchID: branchID,
		},
	})

	return resp.BranchInfo, err
}

// persistence helper
func (s *HistoryPerfSuite) readv2(branch p.HistoryBranch, minID, maxID, lastVersion int64, pageSize int, token []byte) (p.HistoryBranch, []*workflow.HistoryEvent, []byte, error) {

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
func (s *HistoryPerfSuite) appendV2(bi p.HistoryBranch, events []*workflow.HistoryEvent, txnID int64) (p.HistoryBranch, int, error) {

	var resp *p.AppendHistoryNodesResponse
	var err error

	resp, err = s.HistoryMgr.AppendHistoryNodes(&p.AppendHistoryNodesRequest{
		BranchInfo:    bi,
		Events:        events,
		TransactionID: txnID,
		Encoding:      pickRandomEncoding(),
	})

	if err != nil {
		return p.HistoryBranch{}, 0, err
	}
	s.True(resp.Size > 0)

	return resp.BranchInfo, resp.OverrideCount, err
}

// AppendHistoryEvents helper
func (s *HistoryPerfSuite) appendV1(domainID string, workflowExecution workflow.WorkflowExecution,
	firstEventID, eventBatchVersion int64, rangeID, txID int64, eventsBatch *workflow.History, overwrite bool) error {

	_, err := s.HistoryMgr.AppendHistoryEvents(&p.AppendHistoryEventsRequest{
		DomainID:          domainID,
		Execution:         workflowExecution,
		FirstEventID:      firstEventID,
		EventBatchVersion: eventBatchVersion,
		RangeID:           rangeID,
		TransactionID:     txID,
		Events:            eventsBatch.Events,
		Overwrite:         overwrite,
		Encoding:          pickRandomEncoding(),
	})
	return err
}

// GetWorkflowExecutionHistory helper
func (s *HistoryPerfSuite) readV1(domainID string, workflowExecution workflow.WorkflowExecution,
	firstEventID, nextEventID int64, pageSize int, token []byte) (*workflow.History, []byte, error) {

	response, err := s.HistoryMgr.GetWorkflowExecutionHistory(&p.GetWorkflowExecutionHistoryRequest{
		DomainID:      domainID,
		Execution:     workflowExecution,
		FirstEventID:  firstEventID,
		NextEventID:   nextEventID,
		PageSize:      pageSize,
		NextPageToken: token,
	})

	if err != nil {
		return nil, nil, err
	}

	return response.History, response.NextPageToken, nil
}
