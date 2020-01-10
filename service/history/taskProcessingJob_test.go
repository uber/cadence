// Copyright (c) 2020 Uber Technologies, Inc.
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

package history

import (
	"sort"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/task"
)

type (
	taskProcessingJobSuite struct {
		suite.Suite
		*require.Assertions

		controller      *gomock.Controller
		mockDomainCache *cache.MockDomainCache
	}
)

func TestTaskProcessingJobSuite(t *testing.T) {
	s := new(taskProcessingJobSuite)
	suite.Run(t, s)
}

func (s *taskProcessingJobSuite) SetupSuite() {

}

func (s *taskProcessingJobSuite) TearDownSuite() {

}

func (s *taskProcessingJobSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)
}

func (s *taskProcessingJobSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *taskProcessingJobSuite) TestDomainFilter_Filter() {
	testCases := []struct {
		domainIDs      []string
		inverted       bool
		testDomainIDs  []string
		expectedResult []bool
	}{
		{
			domainIDs:      nil,
			inverted:       false,
			testDomainIDs:  []string{"some random domain"},
			expectedResult: []bool{false},
		},
		{
			domainIDs:      []string{"testDomain1", "testDomain2"},
			inverted:       false,
			testDomainIDs:  []string{"testDomain1", "some random domain"},
			expectedResult: []bool{true, false},
		},
		{
			domainIDs:      []string{},
			inverted:       true,
			testDomainIDs:  []string{"any domainID"},
			expectedResult: []bool{true},
		},
		{
			domainIDs:      []string{"testDomain1", "testDomain2"},
			inverted:       true,
			testDomainIDs:  []string{"testDomain1", "some random domain"},
			expectedResult: []bool{false, true},
		},
	}

	for _, tc := range testCases {
		filter := newDomainFilter(tc.domainIDs, tc.inverted)
		for i, testDomain := range tc.testDomainIDs {
			result := filter.filter(testDomain)
			s.Equal(tc.expectedResult[i], result)
		}
	}
}

func (s *taskProcessingJobSuite) TestDomainFilter_Include() {
	testCases := []struct {
		domainIDs         []string
		inverted          bool
		newDomainIDs      []string
		expectedDomainIDs []string
	}{
		{
			domainIDs:         []string{"testDomain1", "testDomain2"},
			inverted:          false,
			newDomainIDs:      []string{"testDomain2", "testDomain3"},
			expectedDomainIDs: []string{"testDomain1", "testDomain2", "testDomain3"},
		},
		{
			domainIDs:         []string{"testDomain1", "testDomain2"},
			inverted:          true,
			newDomainIDs:      []string{"testDomain2", "testDomain3"},
			expectedDomainIDs: []string{"testDomain1"},
		},
	}

	for _, tc := range testCases {
		baseFilter := newDomainFilter(tc.domainIDs, tc.inverted)
		newFilter := baseFilter.include(tc.newDomainIDs)

		// check if the base filter got modified
		expectedBaseFilterDomainIDMap := make(map[string]struct{})
		for _, domainID := range tc.domainIDs {
			expectedBaseFilterDomainIDMap[domainID] = struct{}{}
		}
		s.Equal(expectedBaseFilterDomainIDMap, baseFilter.domainIDs)
		s.Equal(tc.inverted, baseFilter.invertResult)

		expectedNewFilterDomainIDMap := make(map[string]struct{})
		for _, domainID := range tc.expectedDomainIDs {
			expectedNewFilterDomainIDMap[domainID] = struct{}{}
		}
		s.Equal(expectedNewFilterDomainIDMap, newFilter.domainIDs)
		s.Equal(tc.inverted, newFilter.invertResult)
	}
}

func (s *taskProcessingJobSuite) TestDomainFilter_Exclude() {
	testCases := []struct {
		domainIDs         []string
		inverted          bool
		newDomainIDs      []string
		expectedDomainIDs []string
	}{
		{
			domainIDs:         []string{"testDomain1", "testDomain2"},
			inverted:          false,
			newDomainIDs:      []string{"testDomain2", "testDomain3"},
			expectedDomainIDs: []string{"testDomain1"},
		},
		{
			domainIDs:         []string{"testDomain1", "testDomain2"},
			inverted:          true,
			newDomainIDs:      []string{"testDomain2", "testDomain3"},
			expectedDomainIDs: []string{"testDomain1", "testDomain2", "testDomain3"},
		},
	}

	for _, tc := range testCases {
		baseFilter := newDomainFilter(tc.domainIDs, tc.inverted)
		newFilter := baseFilter.exclude(tc.newDomainIDs)

		// check if the base filter got modified
		expectedBaseFilterDomainIDMap := make(map[string]struct{})
		for _, domainID := range tc.domainIDs {
			expectedBaseFilterDomainIDMap[domainID] = struct{}{}
		}
		s.Equal(expectedBaseFilterDomainIDMap, baseFilter.domainIDs)
		s.Equal(tc.inverted, baseFilter.invertResult)

		expectedNewFilterDomainIDMap := make(map[string]struct{})
		for _, domainID := range tc.expectedDomainIDs {
			expectedNewFilterDomainIDMap[domainID] = struct{}{}
		}
		s.Equal(expectedNewFilterDomainIDMap, newFilter.domainIDs)
		s.Equal(tc.inverted, newFilter.invertResult)
	}
}

func (s *taskProcessingJobSuite) TestDomainFilter_Merge() {
	testCases := []struct {
		domainIDs            [][]string
		inverted             []bool
		expectedDomainIDs    []string
		expectedInvertResult bool
	}{
		{
			domainIDs: [][]string{
				{"testDomain1", "testDomain2"},
				{"testDomain2", "testDomain3"},
			},
			inverted:             []bool{false, false},
			expectedDomainIDs:    []string{"testDomain1", "testDomain2", "testDomain3"},
			expectedInvertResult: false,
		},
		{
			domainIDs: [][]string{
				{"testDomain1", "testDomain2"},
				{"testDomain2", "testDomain3"},
			},
			inverted:             []bool{true, true},
			expectedDomainIDs:    []string{"testDomain2"},
			expectedInvertResult: true,
		},
		{
			domainIDs: [][]string{
				{"testDomain1", "testDomain2"},
				{"testDomain2", "testDomain3"},
			},
			inverted:             []bool{true, false},
			expectedDomainIDs:    []string{"testDomain1"},
			expectedInvertResult: true,
		},
		{
			domainIDs: [][]string{
				{"testDomain1", "testDomain2"},
				{"testDomain2", "testDomain3"},
			},
			inverted:             []bool{false, true},
			expectedDomainIDs:    []string{"testDomain3"},
			expectedInvertResult: true,
		},
	}

	for _, tc := range testCases {
		var filters []domainFilter
		for i, domainIDs := range tc.domainIDs {
			filters = append(filters, newDomainFilter(domainIDs, tc.inverted[i]))
		}

		s.NotEmpty(filters)
		mergedFilter := filters[0]
		for _, f := range filters[1:] {
			mergedFilter = mergedFilter.merge(f)
		}

		expectedMergedFilterDomainIDMap := make(map[string]struct{})
		for _, domainID := range tc.expectedDomainIDs {
			expectedMergedFilterDomainIDMap[domainID] = struct{}{}
		}
		s.Equal(expectedMergedFilterDomainIDMap, mergedFilter.domainIDs)
		s.Equal(tc.expectedInvertResult, mergedFilter.invertResult)
	}
}

func (s *taskProcessingJobSuite) TestBaseSplitPolicy() {
	testCases := []struct {
		maxJobLevel                   int
		totalPendingTaskThreshold     int
		perDomainPendingTaskThreshold int
		currentJobLevel               int
		pendingTaskStats              map[string]int
		expectedResult                map[string]int
	}{
		{
			maxJobLevel:                   5,
			totalPendingTaskThreshold:     1000,
			perDomainPendingTaskThreshold: 10,
			currentJobLevel:               0,
			pendingTaskStats:              map[string]int{"testDomain1": 5, "testDomain2": 100},
			expectedResult:                nil,
		},
		{
			maxJobLevel:                   5,
			totalPendingTaskThreshold:     20,
			perDomainPendingTaskThreshold: 10,
			currentJobLevel:               0,
			pendingTaskStats:              map[string]int{"testDomain1": 5, "testDomain2": 100},
			expectedResult:                map[string]int{"testDomain2": 1},
		},
		{
			maxJobLevel:                   5,
			totalPendingTaskThreshold:     20,
			perDomainPendingTaskThreshold: 10,
			currentJobLevel:               5,
			pendingTaskStats:              map[string]int{"testDomain1": 10, "testDomain2": 100},
			expectedResult:                nil,
		},
	}

	for _, tc := range testCases {
		policy := newBaseSplitPolicy(
			tc.maxJobLevel,
			tc.totalPendingTaskThreshold,
			tc.perDomainPendingTaskThreshold,
		)
		result := policy.Evaluate(tc.pendingTaskStats, tc.currentJobLevel)
		s.Equal(tc.expectedResult, result)
	}
}

func (s *taskProcessingJobSuite) TestStandbyDomainSplitPolicy() {
	standbyJobLevel := 100
	currentClusterName := "cluster"
	testCases := []struct {
		pendingTaskStats map[string]int
		currentLevel     int
		isStandbyDomain  map[string]bool
		expectedResult   map[string]int
	}{
		{
			pendingTaskStats: map[string]int{"standbyDomain": 1, "activeDomain": 1000},
			currentLevel:     0,
			isStandbyDomain:  map[string]bool{"standbyDomain": true, "activeDomain": false},
			expectedResult:   map[string]int{"standbyDomain": standbyJobLevel},
		},
		{
			pendingTaskStats: map[string]int{"standbyDomain": 1},
			currentLevel:     standbyJobLevel,
			isStandbyDomain:  map[string]bool{"standbyDomain": true},
			expectedResult:   nil,
		},
	}

	policy := newStandbyDomainSplitPolicy(
		s.mockDomainCache,
		standbyJobLevel,
		currentClusterName,
	)
	for _, tc := range testCases {
		for domainID, isStandby := range tc.isStandbyDomain {
			domainEntry := s.newTestDomainEntry(domainID, isStandby, currentClusterName)
			s.mockDomainCache.EXPECT().GetDomainByID(domainID).Return(domainEntry, nil).AnyTimes()
		}
		result := policy.Evaluate(tc.pendingTaskStats, tc.currentLevel)
		s.Equal(tc.expectedResult, result)
	}
}

func (s *taskProcessingJobSuite) TestAggregatedSplitPolicy() {
	currentLevel := 1
	pendingTaskStats := map[string]int{
		"testDomain1": 1,
		"testDomain2": 100,
		"testDomain3": 1000,
	}
	aggregatedPolicy := newAggregatedSplitPolicy()
	result := aggregatedPolicy.Evaluate(pendingTaskStats, currentLevel)
	s.Empty(result)

	policy1 := NewMocktaskProcessingJobSplitPolicy(s.controller)
	policy1Result := map[string]int{
		"testDomain1": 1,
		"testDomain2": 1,
	}
	policy1.EXPECT().Evaluate(pendingTaskStats, currentLevel).Return(policy1Result)

	policy2 := NewMocktaskProcessingJobSplitPolicy(s.controller)
	policy2Result := map[string]int{
		"testDomain1": 10,
		"testDomain3": 10,
	}
	policy2.EXPECT().Evaluate(pendingTaskStats, currentLevel).Return(policy2Result)

	aggregatedPolicy = newAggregatedSplitPolicy(policy1, policy2)
	expectedResult := map[string]int{
		"testDomain1": 1,
		"testDomain2": 1,
		"testDomain3": 10,
	}
	result = aggregatedPolicy.Evaluate(pendingTaskStats, currentLevel)
	s.Equal(expectedResult, result)
}

func (s *taskProcessingJobSuite) TestCompareTaskKeyLess() {
	now := time.Now()
	testCases := []struct {
		key1 *taskKey
		key2 *taskKey
		less bool
	}{
		{
			key1: &taskKey{
				visibilityTimestamp: now,
				taskID:              10,
			},
			key2: &taskKey{
				visibilityTimestamp: now,
				taskID:              100,
			},
			less: true,
		},
		{
			key1: &taskKey{
				visibilityTimestamp: now,
				taskID:              100,
			},
			key2: &taskKey{
				visibilityTimestamp: now,
				taskID:              100,
			},
			less: false,
		},
		{
			key1: &taskKey{visibilityTimestamp: now},
			key2: &taskKey{visibilityTimestamp: now.Add(time.Second)},
			less: true,
		},
		{
			key1: &taskKey{visibilityTimestamp: now.Add(time.Second)},
			key2: &taskKey{visibilityTimestamp: now},
			less: false,
		},
	}

	for _, tc := range testCases {
		s.Equal(tc.less, compareTaskKeyLess(tc.key1, tc.key2))
	}
}

func (s *taskProcessingJobSuite) TestJob_AddTasks() {
	ackLevel := taskKey{taskID: 1}
	maxReadLevel := taskKey{taskID: 10}
	job := s.newTestTaskProcessingJob(
		0,
		ackLevel,
		ackLevel,
		maxReadLevel,
		make(map[taskKey]queueTask),
		domainFilter{},
		nil,
	)

	taskKeys := []taskKey{
		taskKey{taskID: 0},
		taskKey{taskID: 2},
		taskKey{taskID: 5},
		taskKey{taskID: 20},
	}
	var queueTasks []queueTask
	for _, key := range taskKeys {
		task := NewMockqueueTask(s.controller)
		task.EXPECT().Key().Return(key).AnyTimes()
		queueTasks = append(queueTasks, task)
	}

	job.AddTasks(queueTasks)
	s.Len(job.outstandingTasks, 2)
	s.Equal(taskKey{taskID: 5}, job.readLevel)

	// add the same set of tasks again, should have no effect
	job.AddTasks(queueTasks)
	s.Len(job.outstandingTasks, 2)
	s.Equal(taskKey{taskID: 5}, job.readLevel)
}

func (s *taskProcessingJobSuite) TestJob_UpdateAckLevel() {
	ackLevel := taskKey{taskID: 1}
	maxReadLevel := taskKey{taskID: 10}
	tasks := []struct {
		key   taskKey
		acked bool
	}{
		{
			key:   taskKey{taskID: 1},
			acked: true,
		},
		{
			key:   taskKey{taskID: 2},
			acked: true,
		},
		{
			key:   taskKey{taskID: 5},
			acked: true,
		},
		{
			key:   taskKey{taskID: 7},
			acked: false,
		},
		{
			key:   taskKey{taskID: 9},
			acked: true,
		},
	}

	outstandingTasks := make(map[taskKey]queueTask)
	for _, t := range tasks {
		mockTask := NewMockqueueTask(s.controller)
		taskState := task.TaskStatePending
		if t.acked {
			taskState = task.TaskStateAcked
		}
		mockTask.EXPECT().State().Return(taskState).AnyTimes()
		outstandingTasks[t.key] = mockTask
	}

	job := s.newTestTaskProcessingJob(
		0,
		ackLevel,
		maxReadLevel,
		maxReadLevel,
		outstandingTasks,
		domainFilter{},
		nil,
	)

	job.updateAckLevel()
	s.Equal(taskKey{taskID: 5}, job.ackLevel)
	s.Len(job.outstandingTasks, 2)
}

func (s *taskProcessingJobSuite) TestJob_Merge_NoOverlap() {
	jobLevel := 0
	job1 := s.newTestTaskProcessingJob(
		jobLevel,
		taskKey{taskID: 0},
		taskKey{taskID: 0},
		taskKey{taskID: 10},
		nil,
		domainFilter{},
		nil,
	)
	job2 := s.newTestTaskProcessingJob(
		jobLevel,
		taskKey{taskID: 10},
		taskKey{taskID: 50},
		taskKey{taskID: 100},
		nil,
		domainFilter{},
		nil,
	)
	job3 := s.newTestTaskProcessingJob(
		jobLevel,
		taskKey{taskID: 101},
		taskKey{taskID: 101},
		taskKey{taskID: 1000},
		nil,
		domainFilter{},
		nil,
	)

	s.Nil(job1.Merge(job2))
	s.Nil(job1.Merge(job3))
	s.Nil(job2.Merge(job3))
	s.Nil(job2.Merge(job1))
	s.Nil(job3.Merge(job2))
	s.Nil(job3.Merge(job1))
}

func (s *taskProcessingJobSuite) TestJob_Merge_Overlap() {
	jobLevel := 0
	testCases := []struct {
		testName       string
		job1           *taskProcessingJobImpl
		job2           *taskProcessingJobImpl
		expectedResult []*taskProcessingJobImpl
	}{
		{
			testName: "SameAckLevel",
			job1: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 7},
				taskKey{taskID: 10},
				map[taskKey]queueTask{
					taskKey{taskID: 1}: nil,
					taskKey{taskID: 4}: nil,
					taskKey{taskID: 7}: nil,
				},
				newDomainFilter([]string{"testDomain1"}, false),
				nil,
			),
			job2: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 3},
				taskKey{taskID: 5},
				map[taskKey]queueTask{
					taskKey{taskID: 2}: nil,
					taskKey{taskID: 3}: nil,
				},
				newDomainFilter([]string{"testDomain2"}, false),
				nil,
			),
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 3},
					taskKey{taskID: 5},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil,
						taskKey{taskID: 2}: nil,
						taskKey{taskID: 3}: nil,
						taskKey{taskID: 4}: nil,
					},
					newDomainFilter([]string{"testDomain1", "testDomain2"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 5},
					taskKey{taskID: 7},
					taskKey{taskID: 10},
					map[taskKey]queueTask{
						taskKey{taskID: 7}: nil,
					},
					newDomainFilter([]string{"testDomain1"}, false),
					nil,
				),
			},
		},
		{
			testName: "SameMaxReadLevel",
			job1: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 7},
				taskKey{taskID: 10},
				map[taskKey]queueTask{
					taskKey{taskID: 1}: nil,
					taskKey{taskID: 4}: nil,
					taskKey{taskID: 7}: nil,
				},
				newDomainFilter([]string{"testDomain1"}, false),
				nil,
			),
			job2: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 5},
				taskKey{taskID: 9},
				taskKey{taskID: 10},
				map[taskKey]queueTask{
					taskKey{taskID: 6}: nil,
					taskKey{taskID: 8}: nil,
					taskKey{taskID: 9}: nil,
				},
				newDomainFilter([]string{"testDomain2"}, false),
				nil,
			),
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 5},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil,
						taskKey{taskID: 4}: nil,
					},
					newDomainFilter([]string{"testDomain1"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 5},
					taskKey{taskID: 7},
					taskKey{taskID: 10},
					map[taskKey]queueTask{
						taskKey{taskID: 6}: nil,
						taskKey{taskID: 7}: nil,
						taskKey{taskID: 8}: nil,
						taskKey{taskID: 9}: nil,
					},
					newDomainFilter([]string{"testDomain1", "testDomain2"}, false),
					nil,
				),
			},
		},
		{
			testName: "OneJobContainTheOther",
			job1: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 7},
				taskKey{taskID: 20},
				map[taskKey]queueTask{
					taskKey{taskID: 1}: nil,
					taskKey{taskID: 4}: nil,
					taskKey{taskID: 7}: nil,
				},
				newDomainFilter([]string{"testDomain1"}, false),
				nil,
			),
			job2: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 5},
				taskKey{taskID: 9},
				taskKey{taskID: 10},
				map[taskKey]queueTask{
					taskKey{taskID: 6}: nil,
					taskKey{taskID: 8}: nil,
					taskKey{taskID: 9}: nil,
				},
				newDomainFilter([]string{"testDomain2"}, false),
				nil,
			),
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 5},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil,
						taskKey{taskID: 4}: nil,
					},
					newDomainFilter([]string{"testDomain1"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 5},
					taskKey{taskID: 7},
					taskKey{taskID: 10},
					map[taskKey]queueTask{
						taskKey{taskID: 6}: nil,
						taskKey{taskID: 7}: nil,
						taskKey{taskID: 8}: nil,
						taskKey{taskID: 9}: nil,
					},
					newDomainFilter([]string{"testDomain1", "testDomain2"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 10},
					taskKey{taskID: 10},
					taskKey{taskID: 20},
					map[taskKey]queueTask{},
					newDomainFilter([]string{"testDomain1"}, false),
					nil,
				),
			},
		},
		{
			testName: "GeneralCases",
			job1: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 10},
				taskKey{taskID: 15},
				map[taskKey]queueTask{
					taskKey{taskID: 1}:  nil,
					taskKey{taskID: 4}:  nil,
					taskKey{taskID: 7}:  nil,
					taskKey{taskID: 10}: nil,
				},
				newDomainFilter([]string{"testDomain1"}, false),
				nil,
			),
			job2: s.newTestTaskProcessingJob(
				jobLevel,
				taskKey{taskID: 5},
				taskKey{taskID: 17},
				taskKey{taskID: 20},
				map[taskKey]queueTask{
					taskKey{taskID: 6}:  nil,
					taskKey{taskID: 8}:  nil,
					taskKey{taskID: 9}:  nil,
					taskKey{taskID: 17}: nil,
				},
				newDomainFilter([]string{"testDomain2"}, false),
				nil,
			),
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 5},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil,
						taskKey{taskID: 4}: nil,
					},
					newDomainFilter([]string{"testDomain1"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 5},
					taskKey{taskID: 10},
					taskKey{taskID: 15},
					map[taskKey]queueTask{
						taskKey{taskID: 6}:  nil,
						taskKey{taskID: 7}:  nil,
						taskKey{taskID: 8}:  nil,
						taskKey{taskID: 9}:  nil,
						taskKey{taskID: 10}: nil,
					},
					newDomainFilter([]string{"testDomain1", "testDomain2"}, false),
					nil,
				),
				s.newTestTaskProcessingJob(
					jobLevel,
					taskKey{taskID: 15},
					taskKey{taskID: 17},
					taskKey{taskID: 20},
					map[taskKey]queueTask{
						taskKey{taskID: 17}: nil,
					},
					newDomainFilter([]string{"testDomain2"}, false),
					nil,
				),
			},
		},
	}

	for _, tc := range testCases {
		result := s.copyJob(tc.job1).Merge(s.copyJob(tc.job2))
		s.Equal(len(tc.expectedResult), len(result))
		for i := range result {
			s.assertJobEqual(tc.expectedResult[i], result[i].(*taskProcessingJobImpl))
		}

		result = s.copyJob(tc.job2).Merge(s.copyJob(tc.job1))
		s.Equal(len(tc.expectedResult), len(result))
		for i := range result {
			s.assertJobEqual(tc.expectedResult[i], result[i].(*taskProcessingJobImpl))
		}
	}
}

func (s *taskProcessingJobSuite) TestJob_Split() {
	initialJobLevel := 0
	jobConfig := &taskProcessingJobConfig{
		lookAheadTaskID: 3,
	}
	testCases := []struct {
		testName             string
		job                  *taskProcessingJobImpl
		outstandingTasksInfo map[int]string // taskID to domainID
		expectedPolicyInput  map[string]int
		splitPolicyResult    map[string]int
		expectedResult       []*taskProcessingJobImpl
	}{
		{
			testName: "NoSplitNeeded",
			job: s.newTestTaskProcessingJob(
				initialJobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 3},
				taskKey{taskID: 5},
				nil, // will be assigned using outstandingTasksInfo
				domainFilter{},
				jobConfig,
			),
			outstandingTasksInfo: map[int]string{1: "testDomain1", 2: "testDomain1", 3: "testDomain2"},
			expectedPolicyInput:  map[string]int{"testDomain1": 2, "testDomain2": 1},
			splitPolicyResult:    nil,
			expectedResult:       nil,
		},
		{
			testName: "SplitToOneNewLevel",
			job: s.newTestTaskProcessingJob(
				initialJobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 5},
				taskKey{taskID: 10},
				nil, // will be assigned using outstandingTasksInfo
				newDomainFilter([]string{"testDomain1", "testDomain2", "testDomain3"}, false),
				jobConfig,
			),
			outstandingTasksInfo: map[int]string{1: "testDomain1", 2: "testDomain1", 3: "testDomain2", 5: "testDomain3"},
			expectedPolicyInput:  map[string]int{"testDomain1": 2, "testDomain2": 1, "testDomain3": 1},
			splitPolicyResult:    map[string]int{"testDomain2": 1, "testDomain3": 1},
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					1,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 8},
					map[taskKey]queueTask{
						taskKey{taskID: 3}: nil, // queueTask will be assigned using outstandingTasksInfo
						taskKey{taskID: 5}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain2", "testDomain3"}, false),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					initialJobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 8},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil, // queueTask will be assigned using outstandingTasksInfo
						taskKey{taskID: 2}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain1"}, false),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					initialJobLevel,
					taskKey{taskID: 8},
					taskKey{taskID: 8},
					taskKey{taskID: 10},
					map[taskKey]queueTask{},
					newDomainFilter([]string{"testDomain1", "testDomain2", "testDomain3"}, false),
					jobConfig,
				),
			},
		},
		{
			testName: "SplitToMultipleLevels",
			job: s.newTestTaskProcessingJob(
				initialJobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 5},
				taskKey{taskID: 10},
				nil, // will be assigned using outstandingTasksInfo
				newDomainFilter([]string{}, true),
				jobConfig,
			),
			outstandingTasksInfo: map[int]string{1: "testDomain1", 2: "testDomain1", 3: "testDomain2", 5: "testDomain3"},
			expectedPolicyInput:  map[string]int{"testDomain1": 2, "testDomain2": 1, "testDomain3": 1},
			splitPolicyResult:    map[string]int{"testDomain2": 1, "testDomain3": 2},
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					1,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 8},
					map[taskKey]queueTask{
						taskKey{taskID: 3}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain2"}, false),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					2,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 8},
					map[taskKey]queueTask{
						taskKey{taskID: 5}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain3"}, false),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					initialJobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 8},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil, // queueTask will be assigned using outstandingTasksInfo
						taskKey{taskID: 2}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain2", "testDomain3"}, true),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					initialJobLevel,
					taskKey{taskID: 8},
					taskKey{taskID: 8},
					taskKey{taskID: 10},
					map[taskKey]queueTask{},
					newDomainFilter([]string{}, true),
					jobConfig,
				),
			},
		},
		{
			testName: "LookAheadPassedMaxReadLevel",
			job: s.newTestTaskProcessingJob(
				initialJobLevel,
				taskKey{taskID: 0},
				taskKey{taskID: 5},
				taskKey{taskID: 7},
				nil, // will be assigned using outstandingTasksInfo
				newDomainFilter([]string{}, true),
				jobConfig,
			),
			outstandingTasksInfo: map[int]string{1: "testDomain1", 2: "testDomain1", 3: "testDomain2", 5: "testDomain3"},
			expectedPolicyInput:  map[string]int{"testDomain1": 2, "testDomain2": 1, "testDomain3": 1},
			splitPolicyResult:    map[string]int{"testDomain2": 1},
			expectedResult: []*taskProcessingJobImpl{
				s.newTestTaskProcessingJob(
					1,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 7},
					map[taskKey]queueTask{
						taskKey{taskID: 3}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain2"}, false),
					jobConfig,
				),
				s.newTestTaskProcessingJob(
					initialJobLevel,
					taskKey{taskID: 0},
					taskKey{taskID: 5},
					taskKey{taskID: 7},
					map[taskKey]queueTask{
						taskKey{taskID: 1}: nil, // queueTask will be assigned using outstandingTasksInfo
						taskKey{taskID: 2}: nil, // queueTask will be assigned using outstandingTasksInfo
						taskKey{taskID: 5}: nil, // queueTask will be assigned using outstandingTasksInfo
					},
					newDomainFilter([]string{"testDomain2"}, true),
					jobConfig,
				),
			},
		},
	}

	for _, tc := range testCases {
		mockPolicy := NewMocktaskProcessingJobSplitPolicy(s.controller)
		mockPolicy.EXPECT().Evaluate(tc.expectedPolicyInput, initialJobLevel).Return(tc.splitPolicyResult).Times(1)
		outstandingTasks := make(map[taskKey]queueTask)
		for taskID, domainID := range tc.outstandingTasksInfo {
			mockQueueTaskInfo := NewMockqueueTaskInfo(s.controller)
			mockQueueTaskInfo.EXPECT().GetDomainID().Return(domainID).AnyTimes()
			mockQueueTask := NewMockqueueTask(s.controller)
			mockQueueTask.EXPECT().Info().Return(mockQueueTaskInfo).AnyTimes()
			outstandingTasks[taskKey{taskID: int64(taskID)}] = mockQueueTask
		}
		tc.job.outstandingTasks = outstandingTasks

		for _, expectedJob := range tc.expectedResult {
			for key := range expectedJob.outstandingTasks {
				expectedJob.outstandingTasks[key] = outstandingTasks[key]
			}
		}

		result := tc.job.Split(mockPolicy)
		s.Equal(len(tc.expectedResult), len(result))
		compareJobFunc := func(first, second *taskProcessingJobImpl) bool {
			if first.jobLevel == second.jobLevel {
				return compareTaskKeyLess(&first.ackLevel, &second.ackLevel)
			}
			return first.jobLevel < second.jobLevel
		}
		sort.Slice(tc.expectedResult, func(i, j int) bool {
			return compareJobFunc(tc.expectedResult[i], tc.expectedResult[j])
		})
		sort.Slice(result, func(i, j int) bool {
			return compareJobFunc(result[i].(*taskProcessingJobImpl), result[j].(*taskProcessingJobImpl))
		})
		for i := range tc.expectedResult {
			s.assertJobEqual(tc.expectedResult[i], result[i].(*taskProcessingJobImpl))
		}
	}
}

func (s *taskProcessingJobSuite) assertJobEqual(job1, job2 *taskProcessingJobImpl) {
	s.Equal(job1.jobLevel, job2.jobLevel)
	s.Equal(job1.ackLevel, job2.ackLevel)
	s.Equal(job1.readLevel, job2.readLevel)
	s.Equal(job1.maxReadLevel, job2.maxReadLevel)
	s.Equal(job1.outstandingTasks, job2.outstandingTasks)
	s.Equal(job1.domainFilter.domainIDs, job2.domainFilter.domainIDs)
	s.Equal(job1.domainFilter.invertResult, job2.domainFilter.invertResult)
}

func (s *taskProcessingJobSuite) copyJob(job *taskProcessingJobImpl) *taskProcessingJobImpl {
	outstandingTasks := make(map[taskKey]queueTask)
	for key, task := range job.outstandingTasks {
		outstandingTasks[key] = task
	}
	return s.newTestTaskProcessingJob(
		job.jobLevel,
		job.ackLevel,
		job.readLevel,
		job.maxReadLevel,
		outstandingTasks,
		job.domainFilter.copy(),
		job.config,
	)
}

func (s *taskProcessingJobSuite) newTestTaskProcessingJob(
	jobLevel int,
	ackLevel taskKey,
	readLevel taskKey,
	maxReadLevel taskKey,
	outstandingTasks map[taskKey]queueTask,
	domainFilter domainFilter,
	config *taskProcessingJobConfig,
) *taskProcessingJobImpl {
	return &taskProcessingJobImpl{
		jobLevel:         jobLevel,
		ackLevel:         ackLevel,
		readLevel:        readLevel,
		maxReadLevel:     maxReadLevel,
		outstandingTasks: outstandingTasks,
		domainFilter:     domainFilter,
		config:           config,
	}
}

func (s *taskProcessingJobSuite) newTestDomainEntry(
	domainID string,
	isStandby bool,
	currentClusterName string,
) *cache.DomainCacheEntry {
	activeClusterName := currentClusterName
	if isStandby {
		activeClusterName = "some other cluster name"
	}
	return cache.NewDomainCacheEntryForTest(
		nil, nil, true, &persistence.DomainReplicationConfig{
			ActiveClusterName: activeClusterName,
		}, 0, nil,
	)
}
