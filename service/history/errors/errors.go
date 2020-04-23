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

package errors

import (
	"github.com/uber/cadence/.gen/go/shared"
)

var (
	// ErrRetryEntityNotExists is returned to indicate workflow execution is not created yet and replicator should
	// try this task again after a small delay.
	ErrRetryEntityNotExists = &shared.RetryTaskError{Message: "entity not exists"}
	// ErrRetryRaceCondition is returned to indicate logic race condition encountered and replicator should
	// try this task again after a small delay.
	ErrRetryRaceCondition = &shared.RetryTaskError{Message: "encounter race condition, retry"}
	// ErrRetryBufferEventsMsg is returned when events are arriving out of order, should retry, or specify force apply
	ErrRetryBufferEventsMsg = "retry on applying buffer events"
	// ErrWorkflowNotFoundMsg is returned when workflow not found
	ErrWorkflowNotFoundMsg = "retry on workflow not found"
	// ErrRetryExistingWorkflowMsg is returned when events are arriving out of order, and there is another workflow with same version running
	ErrRetryExistingWorkflowMsg = "workflow with same version is running"
	// ErrRetryExecutionAlreadyStarted is returned to indicate another workflow execution already started,
	// this error can be return if we encounter race condition, i.e. terminating the target workflow while
	// the target workflow has done continue as new.
	// try this task again after a small delay.
	ErrRetryExecutionAlreadyStarted = &shared.RetryTaskError{Message: "another workflow execution is running"}
	// ErrCorruptedReplicationInfo is returned when replication task has corrupted replication information from source cluster
	ErrCorruptedReplicationInfo = &shared.BadRequestError{Message: "replication task is has corrupted cluster replication info"}
	// ErrCorruptedMutableStateDecision is returned when mutable state decision is corrupted
	ErrCorruptedMutableStateDecision = &shared.BadRequestError{Message: "mutable state decision is corrupted"}
	// ErrMoreThan2DC is returned when there are more than 2 data center
	ErrMoreThan2DC = &shared.BadRequestError{Message: "more than 2 data center"}
	// ErrImpossibleLocalRemoteMissingReplicationInfo is returned when replication task is missing replication info, as well as local replication info being empty
	ErrImpossibleLocalRemoteMissingReplicationInfo = &shared.BadRequestError{Message: "local and remote both are missing replication info"}
	// ErrImpossibleRemoteClaimSeenHigherVersion is returned when replication info contains higher version then this cluster ever emitted.
	ErrImpossibleRemoteClaimSeenHigherVersion = &shared.BadRequestError{Message: "replication info contains higher version then this cluster ever emitted"}
	// ErrInternalFailure is returned when encounter code bug
	ErrInternalFailure = &shared.BadRequestError{Message: "fail to apply history events due bug"}
	// ErrEmptyHistoryRawEventBatch indicate that one single batch of history raw events is of size 0
	ErrEmptyHistoryRawEventBatch = &shared.BadRequestError{Message: "encounter empty history batch"}
	// ErrUnknownEncodingType indicate that the encoding type is unknown
	ErrUnknownEncodingType = &shared.BadRequestError{Message: "unknown encoding type"}
	// ErrUnreappliableEvent indicate that the event is not reappliable
	ErrUnreappliableEvent = &shared.BadRequestError{Message: "event is not reappliable"}
	// ErrWorkflowMutationDecision indicate that something is wrong with mutating workflow, i.e. adding decision to workflow
	ErrWorkflowMutationDecision = &shared.BadRequestError{Message: "error encountered when mutating workflow adding decision"}
	// ErrWorkflowMutationSignal indicate that something is wrong with mutating workflow, i.e. adding signal to workflow
	ErrWorkflowMutationSignal = &shared.BadRequestError{Message: "error encountered when mutating workflow adding signal"}
)
