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
	fmt.Printf("%v , time eslapsed: %v milliseconds\n", name, float64(du)/float64(1000000))
}

/*
=== RUN   TestCassandraHistoryPerformance/TestPerf
appendV1-batch size: 1 , time eslapsed: 3543.659 milliseconds
appendV2-batch size: 1 , time eslapsed: 958.665 milliseconds
appendV1-batch size: 5 , time eslapsed: 656.35 milliseconds
appendV2-batch size: 5 , time eslapsed: 197.422 milliseconds
appendV1-batch size: 10 , time eslapsed: 328.444 milliseconds
appendV2-batch size: 10 , time eslapsed: 106.208 milliseconds
appendV1-batch size: 100 , time eslapsed: 46.533 milliseconds
appendV2-batch size: 100 , time eslapsed: 19.485 milliseconds
appendV1-batch size: 1000 , time eslapsed: 16.503 milliseconds
appendV2-batch size: 1000 , time eslapsed: 17.835 milliseconds
readv2-batch size: 1 , time eslapsed: 1856.21 milliseconds
readV1-batch size: 1 , time eslapsed: 998.079 milliseconds
readv2-batch size: 5 , time eslapsed: 220.197 milliseconds
readV1-batch size: 5 , time eslapsed: 216.845 milliseconds
readv2-batch size: 10 , time eslapsed: 109.88 milliseconds
readV1-batch size: 10 , time eslapsed: 112.853 milliseconds
readv2-batch size: 100 , time eslapsed: 15.682 milliseconds
readV1-batch size: 100 , time eslapsed: 17.451 milliseconds
readv2-batch size: 1000 , time eslapsed: 6.693 milliseconds
readV1-batch size: 1000 , time eslapsed: 6.634 milliseconds
time="2018-10-20T16:13:59-07:00" level=info msg="dropped namespace" keyspace=test_owfrwroowo
--- PASS: TestCassandraHistoryPerformance (11.10s)
    --- PASS: TestCassandraHistoryPerformance/TestPerf (9.46s)
PASS
ok  	github.com/uber/cadence/common/persistence/persistence-tests	11.123s
*/
func (s *HistoryPerfSuite) TestPerf() {
	treeID := s.genRandomUUIDString()

	//for v1
	domainID := treeID

	total := 5000
	allBatches := []int{1, 5, 10, 100, 1000}
	//1. test append different batch allBatches of events:
	wfs := [5]workflow.WorkflowExecution{}
	brs := [5][]byte{}

	for idx, batchSize := range allBatches {

		uuid := s.genRandomUUIDString()
		wfs[idx] = workflow.WorkflowExecution{
			WorkflowId: &uuid,
			RunId:      &uuid,
		}

		br, err := s.newHistoryBranch(treeID)
		s.Nil(err)
		s.NotNil(br)
		brs[idx] = br

		firstIDV1 := int64(1)
		firstIDV2 := int64(1)
		st := s.startProfile()
		for i := 0; i < total/batchSize; i++ {
			lastID := firstIDV1 + int64(batchSize)
			events := s.genRandomEvents(firstIDV1, lastID)
			history := &workflow.History{
				Events: events,
			}

			err := s.appendV1(domainID, wfs[idx], firstIDV1, 0, 0, 0, history, false)
			s.Nil(err)
			firstIDV1 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("appendV1-batch size: %v", batchSize))

		st = s.startProfile()
		for i := 0; i < total/batchSize; i++ {
			lastID := firstIDV2 + int64(batchSize)
			events := s.genRandomEvents(firstIDV2, lastID)

			err := s.appendV2(brs[idx], events, 0)

			s.Nil(err)
			firstIDV2 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("appendV2-batch size: %v", batchSize))

	}

	//2. test read events:
	for idx, batchSize := range allBatches {

		firstIDV1 := int64(1)
		firstIDV2 := int64(1)

		st := s.startProfile()

		for i := 0; i < total/batchSize; i++ {
			lastID := firstIDV2 + int64(batchSize)

			events, token, err := s.readv2(brs[idx], firstIDV2, lastID, 0, int(lastID-firstIDV2), nil)

			s.Equal(batchSize, len(events))
			s.Nil(err)
			s.Equal(0, len(token))
			s.Equal(*events[len(events)-1].EventId, lastID-1)
			firstIDV2 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("readv2-batch size: %v", batchSize))

		st = s.startProfile()
		for i := 0; i < total/batchSize; i++ {
			lastID := firstIDV1 + int64(batchSize)

			historyR, token, err := s.readV1(domainID, wfs[idx], firstIDV1, lastID, int(lastID-firstIDV1), nil)
			s.Equal(0, len(token))
			s.Equal(batchSize, len(historyR.Events))
			events := historyR.Events
			s.Nil(err)
			s.Equal(*events[len(events)-1].EventId, lastID-1)
			firstIDV1 = lastID
		}
		s.stopProfile(st, fmt.Sprintf("readV1-batch size: %v", batchSize))

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
func (s *HistoryPerfSuite) newHistoryBranch(treeID string) ([]byte, error) {

	resp, err := s.HistoryV2Mgr.NewHistoryBranch(&p.NewHistoryBranchRequest{
		TreeID: treeID,
	})

	return resp.BranchToken, err
}

// persistence helper
func (s *HistoryPerfSuite) readv2(branch []byte, minID, maxID, lastVersion int64, pageSize int, token []byte) ([]*workflow.HistoryEvent, []byte, error) {

	resp, err := s.HistoryV2Mgr.ReadHistoryBranch(&p.ReadHistoryBranchRequest{
		BranchToken:      branch,
		MinEventID:       minID,
		MaxEventID:       maxID,
		PageSize:         pageSize,
		NextPageToken:    token,
		LastEventVersion: lastVersion,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(resp.History) > 0 {
		s.True(resp.Size > 0)
	}
	return resp.History, resp.NextPageToken, nil
}

// persistence helper
func (s *HistoryPerfSuite) appendV2(br []byte, events []*workflow.HistoryEvent, txnID int64) error {

	var resp *p.AppendHistoryNodesResponse
	var err error

	resp, err = s.HistoryV2Mgr.AppendHistoryNodes(&p.AppendHistoryNodesRequest{
		BranchToken:   br,
		Events:        events,
		TransactionID: txnID,
		Encoding:      common.EncodingTypeThriftRW,
	})
	if err != nil {
		s.True(resp.Size > 0)
	}
	return err
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
		Encoding:          common.EncodingTypeThriftRW,
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
