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

package persistence

import (
	"errors"
	"fmt"
)

type (
	// TimeoutError is returned when a write operation fails due to a timeout
	TimeoutError struct {
		Msg string
	}

	// DBUnavailableError is returned when the database is unavailable, could be for various reasons.
	DBUnavailableError struct {
		Msg string
	}

	// TransactionSizeLimitError is returned when the transaction size is too large
	TransactionSizeLimitError struct {
		Msg string
	}

	// InvalidPersistenceRequestError represents invalid request to persistence
	InvalidPersistenceRequestError struct {
		Msg string
	}

	// CurrentWorkflowConditionFailedError represents a failed conditional update for current workflow record
	CurrentWorkflowConditionFailedError struct {
		Msg string
	}

	// ConditionFailedError represents a failed conditional update for execution record
	ConditionFailedError struct {
		Msg string
	}

	// ShardAlreadyExistError is returned when conditionally creating a shard fails
	ShardAlreadyExistError struct {
		Msg string
	}

	// ShardOwnershipLostError is returned when conditional update fails due to RangeID for the shard
	ShardOwnershipLostError struct {
		ShardID int
		Msg     string
	}

	// WorkflowExecutionAlreadyStartedError is returned when creating a new workflow failed.
	WorkflowExecutionAlreadyStartedError struct {
		Msg              string
		StartRequestID   string
		RunID            string
		State            int
		CloseStatus      int
		LastWriteVersion int64
	}

	DuplicateRequestError struct {
		RequestType WorkflowRequestType
		RunID       string
	}
)

func (e *DuplicateRequestError) Error() string {
	return fmt.Sprintf("Request has already been applied to runID: %s", e.RunID)
}

func (e *InvalidPersistenceRequestError) Error() string {
	return e.Msg
}

func (e *CurrentWorkflowConditionFailedError) Error() string {
	return e.Msg
}

func (e *ConditionFailedError) Error() string {
	return e.Msg
}

func (e *ShardAlreadyExistError) Error() string {
	return e.Msg
}

func (e *ShardOwnershipLostError) Error() string {
	return e.Msg
}

func (e *WorkflowExecutionAlreadyStartedError) Error() string {
	return e.Msg
}

func (e *TimeoutError) Error() string {
	return e.Msg
}

func (e *DBUnavailableError) Error() string {
	return e.Msg
}

func (e *TransactionSizeLimitError) Error() string {
	return e.Msg
}

func AsDuplicateRequestError(err error) (*DuplicateRequestError, bool) {
	var e *DuplicateRequestError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
