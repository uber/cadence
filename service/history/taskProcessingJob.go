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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination taskProcessingJob_mock.go -self_package github.com/uber/cadence/service/history

package history

import (
	"sort"
	"time"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/task"
)

type (
	taskProcessingJob interface {
		Level() int
		AckLevel() taskKey
		ReadLevel() taskKey
		MaxReadLevel() taskKey
		DomainFilter() domainFilter
		AddTasks([]queueTask)
		Split(taskProcessingJobSplitPolicy) []taskProcessingJob
		Merge(taskProcessingJob) []taskProcessingJob
	}

	taskProcessingJobSplitPolicy interface {
		Evaluate(
			pendingTaskStats map[string]int,
			currentLevel int,
		) (domainIDToNewLevel map[string]int)
	}

	taskKey struct {
		taskID              int64
		visibilityTimestamp time.Time
	}

	domainFilter struct {
		domainIDs    map[string]struct{}
		invertResult bool
	}

	baseSplitPolicy struct {
		// TODO: change the following three fields to dynamic configs
		maxJobLevel                   int
		totalPendingTaskThreshold     int
		perDomainPendingTaskThreshold int
	}

	standbyDomainSplitPolicy struct {
		domainCache        cache.DomainCache
		currentClusterName string
		standbyJobLevel    int
	}

	aggregatedSplitPolicy struct {
		policies []taskProcessingJobSplitPolicy
	}

	taskProcessingJobConfig struct {
		lookAheadTaskID       int64
		lookAheadTimeDuration time.Duration
	}

	taskProcessingJobImpl struct {
		jobLevel         int
		ackLevel         taskKey
		readLevel        taskKey
		maxReadLevel     taskKey
		domainFilter     domainFilter
		outstandingTasks map[taskKey]queueTask

		config *taskProcessingJobConfig

		logger        log.Logger
		metricsClient metrics.Client
	}
)

func newTaskProcessingJob(
	jobLevel int,
	ackLevel taskKey,
	maxReadLevel taskKey,
	domainFilter domainFilter,
	config *taskProcessingJobConfig,
	logger log.Logger,
	metricsClient metrics.Client,
) taskProcessingJob {
	return &taskProcessingJobImpl{
		jobLevel:         jobLevel,
		ackLevel:         ackLevel,
		readLevel:        ackLevel,
		maxReadLevel:     maxReadLevel,
		domainFilter:     domainFilter,
		outstandingTasks: make(map[taskKey]queueTask),
		config:           config,
		logger:           logger,
		metricsClient:    metricsClient,
	}
}

func (t *taskProcessingJobImpl) Level() int {
	return t.jobLevel
}

func (t *taskProcessingJobImpl) AckLevel() taskKey {
	t.updateAckLevel()
	return t.ackLevel
}

func (t *taskProcessingJobImpl) ReadLevel() taskKey {
	return t.readLevel
}

func (t *taskProcessingJobImpl) MaxReadLevel() taskKey {
	return t.maxReadLevel
}

func (t *taskProcessingJobImpl) DomainFilter() domainFilter {
	return t.domainFilter
}

func (t *taskProcessingJobImpl) AddTasks(tasks []queueTask) {
	for _, task := range tasks {
		taskKey := task.Key()
		if _, loaded := t.outstandingTasks[taskKey]; loaded {
			// task currently pending
			continue
		}

		if compareTaskKeyLess(&taskKey, &t.ackLevel) {
			continue
		}

		if compareTaskKeyLess(&t.maxReadLevel, &taskKey) {
			continue
		}

		t.outstandingTasks[taskKey] = task
		if compareTaskKeyLess(&t.readLevel, &taskKey) {
			t.readLevel = taskKey
		}
	}
}

func (t *taskProcessingJobImpl) Split(splitPolicy taskProcessingJobSplitPolicy) []taskProcessingJob {
	// step 1: gather statistics for pending tasks, results represented as domainID -> # of pending tasks
	pendingTaskStats := make(map[string]int)
	for _, task := range t.outstandingTasks {
		taskDomainID := task.Info().GetDomainID()
		if _, ok := pendingTaskStats[taskDomainID]; !ok {
			pendingTaskStats[taskDomainID] = 1
		} else {
			pendingTaskStats[taskDomainID] += 1
		}
	}

	// step 2: evaluate split policy
	domainIDToNewJobLevel := splitPolicy.Evaluate(pendingTaskStats, t.Level())
	if domainIDToNewJobLevel == nil || len(domainIDToNewJobLevel) == 0 {
		// not need to split
		return nil
	}

	// step 3: create new jobs based on result of step 2
	newJobsOutstandingTasks := make(map[int]map[taskKey]queueTask)
	remainingTasks := make(map[taskKey]queueTask)
	for key, task := range t.outstandingTasks {
		delete(t.outstandingTasks, key)
		taskDomainID := task.Info().GetDomainID()
		if newLevel, ok := domainIDToNewJobLevel[taskDomainID]; ok {
			if _, ok := newJobsOutstandingTasks[newLevel]; !ok {
				newJobsOutstandingTasks[newLevel] = make(map[taskKey]queueTask)
			}
			newJobsOutstandingTasks[newLevel][key] = task
		} else {
			remainingTasks[key] = task
		}
	}

	var newJobs []taskProcessingJob
	newMaxReadLevel := taskKey{
		taskID:              t.ReadLevel().taskID + t.config.lookAheadTaskID,
		visibilityTimestamp: t.ReadLevel().visibilityTimestamp.Add(t.config.lookAheadTimeDuration),
	}
	if compareTaskKeyLess(&t.maxReadLevel, &newMaxReadLevel) {
		newMaxReadLevel = t.MaxReadLevel()
	}

	newLevelToDomainIDs := make(map[int][]string)
	for domainID, level := range domainIDToNewJobLevel {
		newLevelToDomainIDs[level] = append(newLevelToDomainIDs[level], domainID)
	}
	for level, outstandingTasks := range newJobsOutstandingTasks {
		newJob := &taskProcessingJobImpl{
			jobLevel:         level,
			ackLevel:         t.ackLevel,
			readLevel:        t.readLevel,
			maxReadLevel:     newMaxReadLevel,
			outstandingTasks: outstandingTasks,
			domainFilter:     newDomainFilter(newLevelToDomainIDs[level], false),
			config:           t.config,
			logger:           t.logger,
			metricsClient:    t.metricsClient,
		}
		newJobs = append(newJobs, newJob)
	}

	excludedDomains := []string{}
	for domainID := range domainIDToNewJobLevel {
		excludedDomains = append(excludedDomains, domainID)
	}
	newJob := &taskProcessingJobImpl{
		jobLevel:         t.jobLevel,
		ackLevel:         t.ackLevel,
		readLevel:        t.readLevel,
		maxReadLevel:     newMaxReadLevel,
		outstandingTasks: remainingTasks,
		domainFilter:     t.domainFilter.exclude(excludedDomains),
		config:           t.config,
		logger:           t.logger,
		metricsClient:    t.metricsClient,
	}
	newJobs = append(newJobs, newJob)

	if compareTaskKeyLess(&newMaxReadLevel, &t.maxReadLevel) {
		newJob := newTaskProcessingJob(
			t.jobLevel,
			newMaxReadLevel,
			t.maxReadLevel,
			t.domainFilter.copy(),
			t.config,
			t.logger,
			t.metricsClient,
		)
		newJobs = append(newJobs, newJob)
	}

	return newJobs
}

func (t *taskProcessingJobImpl) Merge(incomingJob taskProcessingJob) []taskProcessingJob {
	incomingJobImpl, ok := incomingJob.(*taskProcessingJobImpl)
	if !ok {
		panic("unknown task processsing job type")
	}

	if t.jobLevel != incomingJobImpl.jobLevel {
		panic("trying to merge job with different job level")
	}

	if !compareTaskKeyLess(&incomingJobImpl.ackLevel, &t.maxReadLevel) {
		return nil
	}

	if !compareTaskKeyLess(&t.ackLevel, &incomingJobImpl.maxReadLevel) {
		return nil
	}

	generationFunc := []func(*taskProcessingJobImpl, *taskProcessingJobImpl) *taskProcessingJobImpl{
		generateMergedPrefixJob,
		generateMergedJob,
		generateMergedSuffixJob,
	}

	mergedJobs := []taskProcessingJob{}
	for _, mergeFn := range generationFunc {
		if mergedJob := mergeFn(t, incomingJobImpl); mergedJob != nil {
			mergedJobs = append(mergedJobs, mergedJob)
		}
	}

	return mergedJobs
}

func generateMergedPrefixJob(
	job1 *taskProcessingJobImpl,
	job2 *taskProcessingJobImpl,
) *taskProcessingJobImpl {
	if compareTaskKeyLess(&job2.ackLevel, &job1.ackLevel) {
		job1, job2 = job2, job1
	}

	if !compareTaskKeyLess(&job1.ackLevel, &job2.ackLevel) {
		// this means job1.ackLevel == job2.ackLevel
		// we don't need to generate a new job in this case
		return nil
	}

	job := &taskProcessingJobImpl{
		jobLevel:         job1.jobLevel,
		ackLevel:         job1.ackLevel,
		readLevel:        minTaskKey(job1.readLevel, job2.ackLevel),
		maxReadLevel:     job2.ackLevel,
		outstandingTasks: make(map[taskKey]queueTask),
		domainFilter:     job1.domainFilter.copy(),
		config:           job1.config,
		logger:           job1.logger,
		metricsClient:    job1.metricsClient,
	}

	for key, task := range job1.outstandingTasks {
		if compareTaskKeyLess(&key, &job2.ackLevel) {
			job.outstandingTasks[key] = task
		}
	}

	return job
}

func generateMergedSuffixJob(
	job1 *taskProcessingJobImpl,
	job2 *taskProcessingJobImpl,
) *taskProcessingJobImpl {
	if compareTaskKeyLess(&job1.maxReadLevel, &job2.maxReadLevel) {
		job1, job2 = job2, job1
	}

	if !compareTaskKeyLess(&job2.maxReadLevel, &job1.maxReadLevel) {
		// this means job1.maxReadLevel == job2.maxReadLevel
		// we don't need to generate a new job in this case
		return nil
	}

	job := &taskProcessingJobImpl{
		jobLevel:         job1.jobLevel,
		ackLevel:         job2.maxReadLevel,
		readLevel:        maxTaskKey(job1.readLevel, job2.maxReadLevel),
		maxReadLevel:     job1.maxReadLevel,
		outstandingTasks: make(map[taskKey]queueTask),
		domainFilter:     job1.domainFilter.copy(),
		config:           job1.config,
		logger:           job1.logger,
		metricsClient:    job1.metricsClient,
	}

	for key, task := range job1.outstandingTasks {
		if !compareTaskKeyLess(&key, &job2.maxReadLevel) {
			job.outstandingTasks[key] = task
		}
	}

	return job
}

func generateMergedJob(
	job1 *taskProcessingJobImpl,
	job2 *taskProcessingJobImpl,
) *taskProcessingJobImpl {
	job := &taskProcessingJobImpl{
		jobLevel:         job1.jobLevel,
		ackLevel:         maxTaskKey(job1.ackLevel, job2.ackLevel),
		readLevel:        minTaskKey(job1.readLevel, job2.readLevel),
		maxReadLevel:     minTaskKey(job1.maxReadLevel, job2.maxReadLevel),
		outstandingTasks: make(map[taskKey]queueTask),
		domainFilter:     job1.domainFilter.merge(job2.domainFilter),
		config:           job1.config,
		logger:           job1.logger,
		metricsClient:    job1.metricsClient,
	}

	for key, task := range job1.outstandingTasks {
		if compareTaskKeyLess(&job2.ackLevel, &key) &&
			!compareTaskKeyLess(&job2.maxReadLevel, &key) {
			job.outstandingTasks[key] = task
		}
	}

	for key, task := range job2.outstandingTasks {
		if compareTaskKeyLess(&job1.ackLevel, &key) &&
			!compareTaskKeyLess(&job1.maxReadLevel, &key) {
			job.outstandingTasks[key] = task
		}
	}

	return job
}

func (t *taskProcessingJobImpl) updateAckLevel() {
	var taskKeys []taskKey
	for key := range t.outstandingTasks {
		taskKeys = append(taskKeys, key)
	}

	sort.Slice(taskKeys, func(i, j int) bool {
		return compareTaskKeyLess(&taskKeys[i], &taskKeys[j])
	})

	for _, currentKey := range taskKeys {
		if t.outstandingTasks[currentKey].State() != task.TaskStateAcked {
			break
		}

		t.ackLevel = currentKey
		delete(t.outstandingTasks, currentKey)
	}
}

func newDomainFilter(
	domainIDs []string,
	inverted bool,
) domainFilter {
	filter := domainFilter{
		domainIDs:    make(map[string]struct{}),
		invertResult: inverted,
	}
	for _, domainID := range domainIDs {
		filter.domainIDs[domainID] = struct{}{}
	}
	return filter
}

func (f *domainFilter) filter(domainID string) bool {
	_, ok := f.domainIDs[domainID]
	if f.invertResult {
		ok = !ok
	}
	return ok
}

func (f *domainFilter) include(domainIDs []string) domainFilter {
	filter := f.copy()
	for _, domainID := range domainIDs {
		if !filter.invertResult {
			filter.domainIDs[domainID] = struct{}{}
		} else {
			delete(filter.domainIDs, domainID)
		}
	}
	return filter
}

func (f *domainFilter) exclude(domainIDs []string) domainFilter {
	filter := f.copy()
	for _, domainID := range domainIDs {
		if !filter.invertResult {
			delete(filter.domainIDs, domainID)
		} else {
			filter.domainIDs[domainID] = struct{}{}
		}
	}
	return filter
}

func (f *domainFilter) merge(f2 domainFilter) domainFilter {
	if !f.invertResult && !f2.invertResult {
		// union the domains
		filter := f.copy()
		for domainID := range f2.domainIDs {
			filter.domainIDs[domainID] = struct{}{}
		}
		return filter
	}

	if f.invertResult && f2.invertResult {
		// intersect the domains
		filter := domainFilter{
			domainIDs:    make(map[string]struct{}),
			invertResult: true,
		}
		for domainID := range f.domainIDs {
			if _, ok := f2.domainIDs[domainID]; ok {
				filter.domainIDs[domainID] = struct{}{}
			}
		}
		return filter
	}

	var filter domainFilter
	var includeDomainIDs map[string]struct{}
	if f.invertResult {
		filter = f.copy()
		includeDomainIDs = f2.domainIDs
	} else {
		filter = f2.copy()
		includeDomainIDs = f.domainIDs
	}

	for domainID := range includeDomainIDs {
		delete(filter.domainIDs, domainID)
	}

	return filter
}

func (f *domainFilter) copy() domainFilter {
	domainIDs := []string{}
	for domainID := range f.domainIDs {
		domainIDs = append(domainIDs, domainID)
	}
	return newDomainFilter(domainIDs, f.invertResult)
}

func newBaseSplitPolicy(
	maxJobLevel int,
	totalPendingTaskThreshold int,
	perDomainPendingTaskThreshold int,
) taskProcessingJobSplitPolicy {
	return &baseSplitPolicy{
		maxJobLevel:                   maxJobLevel,
		totalPendingTaskThreshold:     totalPendingTaskThreshold,
		perDomainPendingTaskThreshold: perDomainPendingTaskThreshold,
	}
}

func (p *baseSplitPolicy) Evaluate(
	pendingTaskStats map[string]int,
	currentLevel int,
) map[string]int {
	if currentLevel >= p.maxJobLevel {
		return nil
	}

	totalPendingTaskNum := 0
	domainIDToNewLevel := make(map[string]int)
	for domainID, numPendingTasks := range pendingTaskStats {
		totalPendingTaskNum += numPendingTasks
		if numPendingTasks > p.perDomainPendingTaskThreshold {
			domainIDToNewLevel[domainID] = currentLevel + 1
		}
	}

	if totalPendingTaskNum < p.totalPendingTaskThreshold {
		return nil
	}
	return domainIDToNewLevel
}

func newStandbyDomainSplitPolicy(
	domainCache cache.DomainCache,
	standbyJobLevel int,
	currentClusterName string,
) taskProcessingJobSplitPolicy {
	return &standbyDomainSplitPolicy{
		domainCache:        domainCache,
		standbyJobLevel:    standbyJobLevel,
		currentClusterName: currentClusterName,
	}
}

func (p *standbyDomainSplitPolicy) Evaluate(
	pendingTaskStats map[string]int,
	currentLevel int,
) map[string]int {
	if currentLevel == p.standbyJobLevel {
		return nil
	}

	domainIDToStandbyLevel := make(map[string]int)
	for domainID := range pendingTaskStats {
		if isStandbyDomain(p.domainCache, domainID, p.currentClusterName) {
			domainIDToStandbyLevel[domainID] = p.standbyJobLevel
		}
	}

	return domainIDToStandbyLevel
}

func newAggregatedSplitPolicy(
	policies ...taskProcessingJobSplitPolicy,
) taskProcessingJobSplitPolicy {
	return &aggregatedSplitPolicy{
		policies: policies,
	}
}

func (p *aggregatedSplitPolicy) Evaluate(
	pendingTaskStats map[string]int,
	currentLevel int,
) map[string]int {
	domainIDToNewLevel := make(map[string]int)

	for _, policy := range p.policies {
		result := policy.Evaluate(pendingTaskStats, currentLevel)
		for domainID, level := range result {
			if _, ok := domainIDToNewLevel[domainID]; !ok {
				domainIDToNewLevel[domainID] = level
			}
		}
	}

	return domainIDToNewLevel
}

func isStandbyDomain(
	domainCache cache.DomainCache,
	domainID string,
	currentClusterName string,
) bool {
	domainEntry, err := domainCache.GetDomainByID(domainID)
	if err != nil {
		// TODO: log the error
		return false
	}
	return domainEntry.IsGlobalDomain() && currentClusterName != domainEntry.GetReplicationConfig().ActiveClusterName
}

func compareTaskKeyLess(first *taskKey, second *taskKey) bool {
	if first.visibilityTimestamp.Equal(second.visibilityTimestamp) {
		return first.taskID < second.taskID
	}
	return first.visibilityTimestamp.Before(second.visibilityTimestamp)
}

func minTaskKey(first taskKey, second taskKey) taskKey {
	if compareTaskKeyLess(&first, &second) {
		return first
	}
	return second
}

func maxTaskKey(first taskKey, second taskKey) taskKey {
	if compareTaskKeyLess(&first, &second) {
		return second
	}
	return first
}
