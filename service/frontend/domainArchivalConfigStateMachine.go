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

package frontend

import (
	"github.com/uber/cadence/.gen/go/shared"
)

type (
	// archivalState represents the state of archival config
	// the only invalid state is enabled=true and bucket=""
	// once bucket is set it is immutable
	// the initial state is enabled=false and bucket="", this is what all domains initially default to
	archivalState struct {
		bucket string
		enabled bool
	}

	// archivalEvent represents a change request to archival config state
	// the only restriction placed on events is that defaultBucket is not empty
	// setting requestedBucket to empty string means user is not attempting to set bucket
	// requestedEnabled can be nil, true, or false (nil indicates no update by user is being attempted)
	archivalEvent struct {
		defaultBucket   string
		bucket string
		enabled *bool
	}
)

// the following errors represents impossible conditions that should never occur
var (
	errInvalidState              = &shared.BadRequestError{Message: "Encountered illegal state: archival is enabled but bucket is not set (should be impossible)."}
	errInvalidEvent           = &shared.BadRequestError{Message: "Encountered illegal event: default bucket is not set (should be impossible)."}
	errCannotHandleStateChange   = &shared.BadRequestError{Message: "Encountered current state and event that cannot be handled (should be impossible)"}
)

// the following errors represents bad user input, but which are possible to occur
var (
	errDisallowedBucketMetadata  = &shared.BadRequestError{Message: "Cannot set bucket owner or bucket retention (must update bucket manually)."}
	errBucketNameUpdate          = &shared.BadRequestError{Message: "Cannot update existing bucket name."}

)

func neverEnabledState() *archivalState {
	return &archivalState{
		bucket: "",
		enabled: false,
	}
}

func registerToEvent(request *shared.RegisterDomainRequest, defaultBucket string) (*archivalEvent, error) {
	event := &archivalEvent{
		defaultBucket:   defaultBucket,
		bucket: request.GetArchivalBucketName(),
		enabled: request.ArchivalEnabled,
	}
	if err := event.validate(); err != nil {
		return nil, err
	}
	return event, nil
}

func updateToEvent(request *shared.UpdateDomainRequest, defaultBucket string) (*archivalEvent, error) {
	event := &archivalEvent{
		defaultBucket: defaultBucket,
	}
	if request.Configuration != nil {
		cfg := request.GetConfiguration()
		if cfg.ArchivalBucketOwner != nil || cfg.ArchivalRetentionPeriodInDays != nil {
			return nil, errDisallowedBucketMetadata
		}
		event.bucket = cfg.GetArchivalBucketName()
		event.enabled = cfg.ArchivalEnabled
	}
	if err := event.validate(); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *archivalEvent) validate() error {
	if len(e.defaultBucket) == 0 {
		return errInvalidEvent
	}
	return nil
}

func (s *archivalState) validate() error {
	if s.enabled && len(s.bucket) == 0 {
		return errInvalidState
	}
	return nil
}

func (s *archivalState) getNextState(e *archivalEvent) (nextState *archivalState, changed bool, err error) {
	defer func() {
		// ensure that any existing bucket name was not mutated
		if nextState != nil && len(s.bucket) != 0 && s.bucket != nextState.bucket {
			nextState = nil
			changed = false
			err = errCannotHandleStateChange
		}

		// ensure that next state is valid
		if nextState != nil {
			if nextStateErr := nextState.validate(); nextStateErr != nil {
				nextState = nil
				changed = false
				err = nextStateErr
			}
		}
	}()

	if s == nil || e == nil {
		return nil, false, errCannotHandleStateChange
	}
	if err := s.validate(); err != nil {
		return nil, false, err
	}
	if err := e.validate(); err != nil {
		return nil, false, err
	}

	/**
	At this point state and event are both non-nil and valid.

	State can be any one of the following:
	{enabled=true,  bucket="foo"}
	{enabled=false, bucket="foo"}
	{enabled=false, bucket=""}

	Event can be any one of the following:
	{enabled=true,  bucket="foo", defaultBucket="bar"}
	{enabled=true,  bucket="",    defaultBucket="bar"}
	{enabled=false, bucket="foo", defaultBucket="bar"}
	{enabled=false, bucket="",    defaultBucket="bar"}
	{enabled=nil,   bucket="foo", defaultBucket="bar"}
	{enabled=nil,   bucket="",    defaultBucket="bar"}
	 */

	 stateBucketSet := len(s.bucket) != 0
	 eventBucketSet := len(e.bucket) != 0

	 // factor this case out to ensure that bucket name is immutable
	 if stateBucketSet && eventBucketSet && s.bucket != e.bucket {
	 	return nil, false, errBucketNameUpdate
	 }

	 // state 1
	 if s.enabled && stateBucketSet {
	 	if e.enabled != nil && *e.enabled && eventBucketSet {
			return s, false, nil
		}
	 	if e.enabled != nil && *e.enabled && !eventBucketSet {
			return s, false, nil
		}
	 	if e.enabled != nil && !(*e.enabled) && eventBucketSet {
			return &archivalState{
				bucket: s.bucket,
				enabled: false,
			}, true, nil
		}
	 	if e.enabled != nil && !(*e.enabled) && !eventBucketSet {
			return &archivalState{
				bucket: s.bucket,
				enabled: false,
			}, true, nil
		}
	 	if e.enabled == nil && eventBucketSet {
			return s, false, nil
		}
	 	if e.enabled == nil && !eventBucketSet {
			return s, false, nil
		}
	 }

	 // state 2
	 if !s.enabled && stateBucketSet {
		 if e.enabled != nil && *e.enabled && eventBucketSet {
		 	return &archivalState{
		 		enabled: true,
		 		bucket: s.bucket,
			}, true, nil
		 }
		 if e.enabled != nil && *e.enabled && !eventBucketSet {
		 	return &archivalState{
		 		enabled: true,
		 		bucket: s.bucket,
			}, true, nil
		 }
		 if e.enabled != nil && !(*e.enabled) && eventBucketSet {
		 	return s, false, nil
		 }
		 if e.enabled != nil && !(*e.enabled) && !eventBucketSet {
			 return s, false, nil
		 }
		 if e.enabled == nil && eventBucketSet {
			return s, false, nil
		 }
		 if e.enabled == nil && !eventBucketSet {
			return s, false, nil
		 }
	 }

	 // state 3
	 if !s.enabled && !stateBucketSet {
		 if e.enabled != nil && *e.enabled && eventBucketSet {
			return &archivalState{
				enabled: true,
				bucket: e.bucket,
			}, true, nil
		 }
		 if e.enabled != nil && *e.enabled && !eventBucketSet {
			return &archivalState{
				enabled: true,
				bucket: e.defaultBucket,
			}, true, nil
		 }
		 if e.enabled != nil && !(*e.enabled) && eventBucketSet {
			return &archivalState{
				enabled: false,
				bucket: e.bucket,
			}, true, nil
		 }
		 if e.enabled != nil && !(*e.enabled) && !eventBucketSet {
			return s, false, nil
		 }
		 if e.enabled == nil && eventBucketSet {
			return &archivalState{
				enabled: false,
				bucket: e.bucket,
			}, true, nil
		 }
		 if e.enabled == nil && !eventBucketSet {
			return s, false, nil
		 }
	 }

	 return nil, false, errCannotHandleStateChange
}
