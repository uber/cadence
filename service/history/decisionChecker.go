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

package history

import (
	"fmt"
	"strings"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/elasticsearch/validator"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
)

type (
	decisionAttrValidator struct {
		config                    *config.Config
		domainCache               cache.DomainCache
		maxIDLengthLimit          int
		searchAttributesValidator *validator.SearchAttributesValidator
	}

	workflowSizeChecker struct {
		blobSizeLimitWarn  int
		blobSizeLimitError int

		historySizeLimitWarn  int
		historySizeLimitError int

		historyCountLimitWarn  int
		historyCountLimitError int

		completedID    int64
		mutableState   execution.MutableState
		executionStats *persistence.ExecutionStats
		metricsScope   metrics.Scope
		logger         log.Logger
	}
)

func newDecisionAttrValidator(
	domainCache cache.DomainCache,
	config *config.Config,
	logger log.Logger,
) *decisionAttrValidator {
	return &decisionAttrValidator{
		config:           config,
		domainCache:      domainCache,
		maxIDLengthLimit: config.MaxIDLengthLimit(),
		searchAttributesValidator: validator.NewSearchAttributesValidator(
			logger,
			config.ValidSearchAttributes,
			config.SearchAttributesNumberOfKeysLimit,
			config.SearchAttributesSizeOfValueLimit,
			config.SearchAttributesTotalSizeLimit,
		),
	}
}

func newWorkflowSizeChecker(
	blobSizeLimitWarn int,
	blobSizeLimitError int,
	historySizeLimitWarn int,
	historySizeLimitError int,
	historyCountLimitWarn int,
	historyCountLimitError int,
	completedID int64,
	mutableState execution.MutableState,
	executionStats *persistence.ExecutionStats,
	metricsScope metrics.Scope,
	logger log.Logger,
) *workflowSizeChecker {
	return &workflowSizeChecker{
		blobSizeLimitWarn:      blobSizeLimitWarn,
		blobSizeLimitError:     blobSizeLimitError,
		historySizeLimitWarn:   historySizeLimitWarn,
		historySizeLimitError:  historySizeLimitError,
		historyCountLimitWarn:  historyCountLimitWarn,
		historyCountLimitError: historyCountLimitError,
		completedID:            completedID,
		mutableState:           mutableState,
		executionStats:         executionStats,
		metricsScope:           metricsScope,
		logger:                 logger,
	}
}

func (c *workflowSizeChecker) failWorkflowIfBlobSizeExceedsLimit(
	decisionTypeTag metrics.Tag,
	blob []byte,
	message string,
) (bool, error) {

	executionInfo := c.mutableState.GetExecutionInfo()
	err := common.CheckEventBlobSizeLimit(
		len(blob),
		c.blobSizeLimitWarn,
		c.blobSizeLimitError,
		executionInfo.DomainID,
		executionInfo.WorkflowID,
		executionInfo.RunID,
		c.metricsScope.Tagged(decisionTypeTag),
		c.logger,
		tag.BlobSizeViolationOperation(decisionTypeTag.Value()),
	)
	if err == nil {
		return false, nil
	}

	attributes := &types.FailWorkflowExecutionDecisionAttributes{
		Reason:  common.StringPtr(common.FailureReasonDecisionBlobSizeExceedsLimit),
		Details: []byte(message),
	}

	if _, err := c.mutableState.AddFailWorkflowEvent(c.completedID, attributes); err != nil {
		return false, err
	}

	return true, nil
}

func (c *workflowSizeChecker) failWorkflowSizeExceedsLimit() (bool, error) {
	historyCount := int(c.mutableState.GetNextEventID()) - 1
	historySize := int(c.executionStats.HistorySize)

	if historySize > c.historySizeLimitError || historyCount > c.historyCountLimitError {
		executionInfo := c.mutableState.GetExecutionInfo()
		c.logger.Error("history size exceeds error limit.",
			tag.WorkflowDomainID(executionInfo.DomainID),
			tag.WorkflowID(executionInfo.WorkflowID),
			tag.WorkflowRunID(executionInfo.RunID),
			tag.WorkflowHistorySize(historySize),
			tag.WorkflowEventCount(historyCount))

		attributes := &types.FailWorkflowExecutionDecisionAttributes{
			Reason:  common.StringPtr(common.FailureReasonSizeExceedsLimit),
			Details: []byte("Workflow history size / count exceeds limit."),
		}

		if _, err := c.mutableState.AddFailWorkflowEvent(c.completedID, attributes); err != nil {
			return false, err
		}
		return true, nil
	}

	if historySize > c.historySizeLimitWarn || historyCount > c.historyCountLimitWarn {
		executionInfo := c.mutableState.GetExecutionInfo()
		c.logger.Warn("history size exceeds warn limit.",
			tag.WorkflowDomainID(executionInfo.DomainID),
			tag.WorkflowID(executionInfo.WorkflowID),
			tag.WorkflowRunID(executionInfo.RunID),
			tag.WorkflowHistorySize(historySize),
			tag.WorkflowEventCount(historyCount))
		return false, nil
	}

	return false, nil
}

func (v *decisionAttrValidator) validateActivityScheduleAttributes(
	domainID string,
	targetDomainID string,
	attributes *types.ScheduleActivityTaskDecisionAttributes,
	wfTimeout int32,
) error {

	if err := v.validateCrossDomainCall(
		domainID,
		targetDomainID,
	); err != nil {
		return err
	}

	if attributes == nil {
		return &types.BadRequestError{Message: "ScheduleActivityTaskDecisionAttributes is not set on decision."}
	}

	defaultTaskListName := ""
	if _, err := v.validatedTaskList(attributes.TaskList, defaultTaskListName); err != nil {
		return err
	}

	if attributes.GetActivityID() == "" {
		return &types.BadRequestError{Message: "ActivityId is not set on decision."}
	}

	if attributes.ActivityType == nil || attributes.ActivityType.GetName() == "" {
		return &types.BadRequestError{Message: "ActivityType is not set on decision."}
	}

	if err := common.ValidateRetryPolicy(attributes.RetryPolicy); err != nil {
		return err
	}

	if len(attributes.GetActivityID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "ActivityID exceeds length limit."}
	}

	if len(attributes.GetActivityType().GetName()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "ActivityType exceeds length limit."}
	}

	if len(attributes.GetDomain()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "Domain exceeds length limit."}
	}

	// Only attempt to deduce and fill in unspecified timeouts only when all timeouts are non-negative.
	if attributes.GetScheduleToCloseTimeoutSeconds() < 0 || attributes.GetScheduleToStartTimeoutSeconds() < 0 ||
		attributes.GetStartToCloseTimeoutSeconds() < 0 || attributes.GetHeartbeatTimeoutSeconds() < 0 {
		return &types.BadRequestError{Message: "A valid timeout may not be negative."}
	}

	// ensure activity timeout never larger than workflow timeout
	if attributes.GetScheduleToCloseTimeoutSeconds() > wfTimeout {
		attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(wfTimeout)
	}
	if attributes.GetScheduleToStartTimeoutSeconds() > wfTimeout {
		attributes.ScheduleToStartTimeoutSeconds = common.Int32Ptr(wfTimeout)
	}
	if attributes.GetStartToCloseTimeoutSeconds() > wfTimeout {
		attributes.StartToCloseTimeoutSeconds = common.Int32Ptr(wfTimeout)
	}
	if attributes.GetHeartbeatTimeoutSeconds() > wfTimeout {
		attributes.HeartbeatTimeoutSeconds = common.Int32Ptr(wfTimeout)
	}

	validScheduleToClose := attributes.GetScheduleToCloseTimeoutSeconds() > 0
	validScheduleToStart := attributes.GetScheduleToStartTimeoutSeconds() > 0
	validStartToClose := attributes.GetStartToCloseTimeoutSeconds() > 0

	if validScheduleToClose {
		if !validScheduleToStart {
			attributes.ScheduleToStartTimeoutSeconds = common.Int32Ptr(attributes.GetScheduleToCloseTimeoutSeconds())
		}
		if !validStartToClose {
			attributes.StartToCloseTimeoutSeconds = common.Int32Ptr(attributes.GetScheduleToCloseTimeoutSeconds())
		}
	} else if validScheduleToStart && validStartToClose {
		attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(attributes.GetScheduleToStartTimeoutSeconds() + attributes.GetStartToCloseTimeoutSeconds())
		if attributes.GetScheduleToCloseTimeoutSeconds() > wfTimeout {
			attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(wfTimeout)
		}
	} else {
		// Deduction failed as there's not enough information to fill in missing timeouts.
		return &types.BadRequestError{Message: "A valid ScheduleToCloseTimeout is not set on decision."}
	}

	// ensure activity's SCHEDULE_TO_START and SCHEDULE_TO_CLOSE is as long as expiration on retry policy
	// if SCHEDULE_TO_START timeout is retryable
	p := attributes.RetryPolicy
	if p != nil {
		isScheduleToStartRetryable := true
		scheduleToStartErrorReason := execution.TimerTypeToReason(execution.TimerTypeScheduleToStart)
		for _, reason := range p.GetNonRetriableErrorReasons() {
			if reason == scheduleToStartErrorReason {
				isScheduleToStartRetryable = false
				break
			}
		}

		expiration := p.GetExpirationIntervalInSeconds()
		if expiration == 0 || expiration > wfTimeout {
			expiration = wfTimeout
		}

		if isScheduleToStartRetryable {
			// If schedule to start timeout is retryable, we don't need to fail the activity and schedule
			// it again on the same tasklist, as it's a no-op). Extending schedule to start timeout to achieve
			// the same thing.
			//
			// Theoretically, we can extend schedule to start to be as long as the expiration time,
			// but if user specifies a very long expiration time and activity task got lost after the activity is
			// scheduled, workflow will be stuck for a long time. So here, we cap the schedule to start timeout
			// to a maximum value, so that when the activity task got lost, timeout can happen sooner and schedule
			// the activity again.

			domainName, _ := v.domainCache.GetDomainName(domainID) // if this call returns an error, we will just used the default value for max timeout
			maximumScheduleToStartTimeoutForRetryInSeconds := int32(v.config.ActivityMaxScheduleToStartTimeoutForRetry(domainName).Seconds())
			scheduleToStartExpiration := common.MinInt32(expiration, maximumScheduleToStartTimeoutForRetryInSeconds)
			if attributes.GetScheduleToStartTimeoutSeconds() < scheduleToStartExpiration {
				attributes.ScheduleToStartTimeoutSeconds = common.Int32Ptr(scheduleToStartExpiration)
			}

			// TODO: uncomment the following code when the client side bug for calculating scheduleToClose deadline is fixed and
			// fully rolled out. Before that, we still need to extend scheduleToClose timeout to be as long as the expiration interval
			//
			// scheduleToCloseExpiration := common.MinInt32(expiration, scheduleToStartExpiration+attributes.GetStartToCloseTimeoutSeconds())
			// if attributes.GetScheduleToCloseTimeoutSeconds() < scheduleToCloseExpiration {
			// 	attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(scheduleToCloseExpiration)
			// }
		}

		if attributes.GetScheduleToCloseTimeoutSeconds() < expiration {
			attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(expiration)
		}
	}
	return nil
}

func (v *decisionAttrValidator) validateTimerScheduleAttributes(
	attributes *types.StartTimerDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "StartTimerDecisionAttributes is not set on decision."}
	}
	if attributes.GetTimerID() == "" {
		return &types.BadRequestError{Message: "TimerId is not set on decision."}
	}
	if len(attributes.GetTimerID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "TimerId exceeds length limit."}
	}
	if attributes.GetStartToFireTimeoutSeconds() <= 0 {
		return &types.BadRequestError{Message: "A valid StartToFireTimeoutSeconds is not set on decision."}
	}
	return nil
}

func (v *decisionAttrValidator) validateActivityCancelAttributes(
	attributes *types.RequestCancelActivityTaskDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "RequestCancelActivityTaskDecisionAttributes is not set on decision."}
	}
	if attributes.GetActivityID() == "" {
		return &types.BadRequestError{Message: "ActivityId is not set on decision."}
	}
	if len(attributes.GetActivityID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "ActivityId exceeds length limit."}
	}
	return nil
}

func (v *decisionAttrValidator) validateTimerCancelAttributes(
	attributes *types.CancelTimerDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "CancelTimerDecisionAttributes is not set on decision."}
	}
	if attributes.GetTimerID() == "" {
		return &types.BadRequestError{Message: "TimerId is not set on decision."}
	}
	if len(attributes.GetTimerID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "TimerId exceeds length limit."}
	}
	return nil
}

func (v *decisionAttrValidator) validateRecordMarkerAttributes(
	attributes *types.RecordMarkerDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "RecordMarkerDecisionAttributes is not set on decision."}
	}
	if attributes.GetMarkerName() == "" {
		return &types.BadRequestError{Message: "MarkerName is not set on decision."}
	}
	if len(attributes.GetMarkerName()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "MarkerName exceeds length limit."}
	}

	return nil
}

func (v *decisionAttrValidator) validateCompleteWorkflowExecutionAttributes(
	attributes *types.CompleteWorkflowExecutionDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "CompleteWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	return nil
}

func (v *decisionAttrValidator) validateFailWorkflowExecutionAttributes(
	attributes *types.FailWorkflowExecutionDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "FailWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	if attributes.Reason == nil {
		return &types.BadRequestError{Message: "Reason is not set on decision."}
	}
	return nil
}

func (v *decisionAttrValidator) validateCancelWorkflowExecutionAttributes(
	attributes *types.CancelWorkflowExecutionDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "CancelWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	return nil
}

func (v *decisionAttrValidator) validateCancelExternalWorkflowExecutionAttributes(
	domainID string,
	targetDomainID string,
	attributes *types.RequestCancelExternalWorkflowExecutionDecisionAttributes,
) error {

	if err := v.validateCrossDomainCall(
		domainID,
		targetDomainID,
	); err != nil {
		return err
	}

	if attributes == nil {
		return &types.BadRequestError{Message: "RequestCancelExternalWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	if attributes.WorkflowID == nil {
		return &types.BadRequestError{Message: "WorkflowId is not set on decision."}
	}
	if len(attributes.GetDomain()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "Domain exceeds length limit."}
	}
	if len(attributes.GetWorkflowID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "WorkflowId exceeds length limit."}
	}
	runID := attributes.GetRunID()
	if runID != "" && uuid.Parse(runID) == nil {
		return &types.BadRequestError{Message: "Invalid RunId set on decision."}
	}

	return nil
}

func (v *decisionAttrValidator) validateSignalExternalWorkflowExecutionAttributes(
	domainID string,
	targetDomainID string,
	attributes *types.SignalExternalWorkflowExecutionDecisionAttributes,
) error {

	if err := v.validateCrossDomainCall(
		domainID,
		targetDomainID,
	); err != nil {
		return err
	}

	if attributes == nil {
		return &types.BadRequestError{Message: "SignalExternalWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	if attributes.Execution == nil {
		return &types.BadRequestError{Message: "Execution is nil on decision."}
	}
	if attributes.Execution.WorkflowID == nil {
		return &types.BadRequestError{Message: "WorkflowId is not set on decision."}
	}
	if len(attributes.GetDomain()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "Domain exceeds length limit."}
	}
	if len(attributes.Execution.GetWorkflowID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "WorkflowId exceeds length limit."}
	}

	targetRunID := attributes.Execution.GetRunID()
	if targetRunID != "" && uuid.Parse(targetRunID) == nil {
		return &types.BadRequestError{Message: "Invalid RunId set on decision."}
	}
	if attributes.SignalName == nil {
		return &types.BadRequestError{Message: "SignalName is not set on decision."}
	}

	return nil
}

func (v *decisionAttrValidator) validateUpsertWorkflowSearchAttributes(
	domainName string,
	attributes *types.UpsertWorkflowSearchAttributesDecisionAttributes,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "UpsertWorkflowSearchAttributesDecisionAttributes is not set on decision."}
	}

	if attributes.SearchAttributes == nil {
		return &types.BadRequestError{Message: "SearchAttributes is not set on decision."}
	}

	if len(attributes.GetSearchAttributes().GetIndexedFields()) == 0 {
		return &types.BadRequestError{Message: "IndexedFields is empty on decision."}
	}

	return v.searchAttributesValidator.ValidateSearchAttributes(attributes.GetSearchAttributes(), domainName)
}

func (v *decisionAttrValidator) validateContinueAsNewWorkflowExecutionAttributes(
	attributes *types.ContinueAsNewWorkflowExecutionDecisionAttributes,
	executionInfo *persistence.WorkflowExecutionInfo,
) error {

	if attributes == nil {
		return &types.BadRequestError{Message: "ContinueAsNewWorkflowExecutionDecisionAttributes is not set on decision."}
	}

	// Inherit workflow type from previous execution if not provided on decision
	if attributes.WorkflowType == nil || attributes.WorkflowType.GetName() == "" {
		attributes.WorkflowType = &types.WorkflowType{Name: common.StringPtr(executionInfo.WorkflowTypeName)}
	}

	if len(attributes.WorkflowType.GetName()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "WorkflowType exceeds length limit."}
	}

	// Inherit Tasklist from previous execution if not provided on decision
	taskList, err := v.validatedTaskList(attributes.TaskList, executionInfo.TaskList)
	if err != nil {
		return err
	}
	attributes.TaskList = taskList

	// Inherit workflow timeout from previous execution if not provided on decision
	if attributes.GetExecutionStartToCloseTimeoutSeconds() <= 0 {
		attributes.ExecutionStartToCloseTimeoutSeconds = common.Int32Ptr(executionInfo.WorkflowTimeout)
	}

	// Inherit decision task timeout from previous execution if not provided on decision
	if attributes.GetTaskStartToCloseTimeoutSeconds() <= 0 {
		attributes.TaskStartToCloseTimeoutSeconds = common.Int32Ptr(executionInfo.DecisionStartToCloseTimeout)
	}

	// Check next run decision task delay
	if attributes.GetBackoffStartIntervalInSeconds() < 0 {
		return &types.BadRequestError{Message: "BackoffStartInterval is less than 0."}
	}

	domainEntry, err := v.domainCache.GetDomainByID(executionInfo.DomainID)
	if err != nil {
		return err
	}
	return v.searchAttributesValidator.ValidateSearchAttributes(attributes.GetSearchAttributes(), domainEntry.GetInfo().Name)
}

func (v *decisionAttrValidator) validateStartChildExecutionAttributes(
	domainID string,
	targetDomainID string,
	attributes *types.StartChildWorkflowExecutionDecisionAttributes,
	parentInfo *persistence.WorkflowExecutionInfo,
) error {

	if err := v.validateCrossDomainCall(
		domainID,
		targetDomainID,
	); err != nil {
		return err
	}

	if attributes == nil {
		return &types.BadRequestError{Message: "StartChildWorkflowExecutionDecisionAttributes is not set on decision."}
	}

	if attributes.GetWorkflowID() == "" {
		return &types.BadRequestError{Message: "Required field WorkflowID is not set on decision."}
	}

	if attributes.WorkflowType == nil || attributes.WorkflowType.GetName() == "" {
		return &types.BadRequestError{Message: "Required field WorkflowType is not set on decision."}
	}

	if len(attributes.GetDomain()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "Domain exceeds length limit."}
	}

	if len(attributes.GetWorkflowID()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "WorkflowId exceeds length limit."}
	}

	if len(attributes.WorkflowType.GetName()) > v.maxIDLengthLimit {
		return &types.BadRequestError{Message: "WorkflowType exceeds length limit."}
	}

	if err := common.ValidateRetryPolicy(attributes.RetryPolicy); err != nil {
		return err
	}

	if err := backoff.ValidateSchedule(attributes.GetCronSchedule()); err != nil {
		return err
	}

	// Inherit tasklist from parent workflow execution if not provided on decision
	taskList, err := v.validatedTaskList(attributes.TaskList, parentInfo.TaskList)
	if err != nil {
		return err
	}
	attributes.TaskList = taskList

	// Inherit workflow timeout from parent workflow execution if not provided on decision
	if attributes.GetExecutionStartToCloseTimeoutSeconds() <= 0 {
		attributes.ExecutionStartToCloseTimeoutSeconds = common.Int32Ptr(parentInfo.WorkflowTimeout)
	}

	// Inherit decision task timeout from parent workflow execution if not provided on decision
	if attributes.GetTaskStartToCloseTimeoutSeconds() <= 0 {
		attributes.TaskStartToCloseTimeoutSeconds = common.Int32Ptr(parentInfo.DecisionStartToCloseTimeout)
	}

	return nil
}

func (v *decisionAttrValidator) validatedTaskList(
	taskList *types.TaskList,
	defaultVal string,
) (*types.TaskList, error) {

	if taskList == nil {
		taskList = &types.TaskList{}
	}

	if taskList.GetName() == "" {
		if defaultVal == "" {
			return taskList, &types.BadRequestError{Message: "missing task list name"}
		}
		taskList.Name = &defaultVal
		return taskList, nil
	}

	name := taskList.GetName()
	if len(name) > v.maxIDLengthLimit {
		return taskList, &types.BadRequestError{
			Message: fmt.Sprintf("task list name exceeds length limit of %v", v.maxIDLengthLimit),
		}
	}

	if strings.HasPrefix(name, common.ReservedTaskListPrefix) {
		return taskList, &types.BadRequestError{
			Message: fmt.Sprintf("task list name cannot start with reserved prefix %v", common.ReservedTaskListPrefix),
		}
	}

	return taskList, nil
}

func (v *decisionAttrValidator) validateCrossDomainCall(
	domainID string,
	targetDomainID string,
) error {

	// same name, no check needed
	if domainID == targetDomainID {
		return nil
	}

	domainEntry, err := v.domainCache.GetDomainByID(domainID)
	if err != nil {
		return err
	}

	targetDomainEntry, err := v.domainCache.GetDomainByID(targetDomainID)
	if err != nil {
		return err
	}

	// both local domain
	if !domainEntry.IsGlobalDomain() && !targetDomainEntry.IsGlobalDomain() {
		return nil
	}

	domainClusters := domainEntry.GetReplicationConfig().Clusters
	targetDomainClusters := targetDomainEntry.GetReplicationConfig().Clusters

	// one is local domain, another one is global domain or both global domain
	// treat global domain with one replication cluster as local domain
	if len(domainClusters) == 1 && len(targetDomainClusters) == 1 {
		if *domainClusters[0] == *targetDomainClusters[0] {
			return nil
		}
		return v.createCrossDomainCallError(domainEntry, targetDomainEntry)
	}
	return v.createCrossDomainCallError(domainEntry, targetDomainEntry)
}

func (v *decisionAttrValidator) createCrossDomainCallError(
	domainEntry *cache.DomainCacheEntry,
	targetDomainEntry *cache.DomainCacheEntry,
) error {
	return &types.BadRequestError{Message: fmt.Sprintf(
		"cannot make cross domain call between %v and %v",
		domainEntry.GetInfo().Name,
		targetDomainEntry.GetInfo().Name,
	)}
}
