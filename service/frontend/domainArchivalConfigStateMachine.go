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
	"context"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
)

// domainArchivalConfigStateMachine is only used by workflowHandler.
// It is simply meant to simplify the logic around archival domain state changes.
// Logically this class can be thought of as part of workflowHandler.

type (
	// archivalState represents the state of archival config
	// the only invalid state is {URI="", status=enabled}
	// once URI is set it is immutable
	// the initial state is {URI="", status=disabled}, this is what all domains initially default to
	archivalState struct {
		historyStatus    shared.ArchivalStatus
		historyURI       string
		visibilityStatus shared.ArchivalStatus
		visibilityURI    string
	}

	// archivalEvent represents a change request to archival config state
	// the only restriction placed on events is that defaultURI is not empty
	// setting requestedURI to empty string means user is not attempting to set URI
	// status can be nil, enabled, or disabled (nil indicates no update by user is being attempted)
	archivalEvent struct {
		defaultHistoryURI    string
		historyURI           string
		historyStatus        *shared.ArchivalStatus
		defaultVisibilityURI string
		visibilityURI        string
		visibilityStatus     *shared.ArchivalStatus
	}
)

// the following errors represents impossible code states that should never occur
var (
	errInvalidState            = &shared.BadRequestError{Message: "Encountered illegal state: archival is enabled but URI is not set (should be impossible)"}
	errInvalidEvent            = &shared.BadRequestError{Message: "Encountered illegal event: default URI is not set (should be impossible)"}
	errCannotHandleStateChange = &shared.BadRequestError{Message: "Encountered current state and event that cannot be handled (should be impossible)"}
)

// the following errors represents bad user input
var (
	errURIUpdate  = &shared.BadRequestError{Message: "Cannot update existing archival URI"}
	errInvalidURI = &shared.BadRequestError{Message: "Invalid archival URI"}
)

func neverEnabledState() *archivalState {
	return &archivalState{
		historyURI:       "",
		historyStatus:    shared.ArchivalStatusDisabled,
		visibilityURI:    "",
		visibilityStatus: shared.ArchivalStatusDisabled,
	}
}

func (e *archivalEvent) validate() error {
	if len(e.defaultHistoryURI) == 0 || len(e.defaultVisibilityURI) == 0 {
		return errInvalidEvent
	}
	return nil
}

func (s *archivalState) validate() error {
	historyStateInvalid := s.historyStatus == shared.ArchivalStatusEnabled && len(s.historyURI) == 0
	visibilityStateInvalid := s.visibilityStatus == shared.ArchivalStatusEnabled && len(s.visibilityURI) == 0
	if historyStateInvalid || visibilityStateInvalid {
		return errInvalidState
	}
	return nil
}

func (s *archivalState) getNextState(
	ctx context.Context,
	historyArchiver archiver.HistoryArchiver,
	visibilityArchiver archiver.VisibilityArchiver,
	e *archivalEvent,
) (nextState *archivalState, changed bool, err error) {
	defer func() {
		// ensure that any existing URI name was not mutated
		historyURIChanged := nextState != nil && len(s.historyURI) != 0 && s.historyURI != nextState.historyURI
		visibilityURIChanged := nextState != nil && len(s.visibilityURI) != 0 && s.visibilityURI != nextState.visibilityURI
		if historyURIChanged || visibilityURIChanged {
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

		if nextState != nil && nextState.historyURI != "" {
			if err = historyArchiver.ValidateURI(nextState.historyURI); err != nil {
				nextState = nil
				changed = false
				err = errInvalidURI
			}
		}

		if nextState != nil && nextState.visibilityURI != "" {
			if err = visibilityArchiver.ValidateURI(nextState.visibilityURI); err != nil {
				nextState = nil
				changed = false
				err = errInvalidURI
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

	nextHistoryStatus, nextHistoryURI, historyStateChanged, err := getNextStateHelper(s.historyStatus, s.historyURI, e.historyStatus, e.historyURI, e.defaultHistoryURI)
	if err != nil {
		return nil, false, err
	}
	nextState.historyStatus = *nextHistoryStatus
	nextState.historyURI = *nextHistoryURI

	nextVisibilityStatus, nextVisibilityURI, visibilityStateChanged, err := getNextStateHelper(s.visibilityStatus, s.visibilityURI, e.visibilityStatus, e.visibilityURI, e.defaultVisibilityURI)
	if err != nil {
		return nil, false, err
	}
	nextState.visibilityStatus = *nextVisibilityStatus
	nextState.visibilityURI = *nextVisibilityURI
	return nextState, historyStateChanged || visibilityStateChanged, nil
}

func getNextStateHelper(
	stateStatus shared.ArchivalStatus,
	stateURI string,
	eventStatus *shared.ArchivalStatus,
	eventURI string,
	eventDefaultURI string,
) (nextStatus *shared.ArchivalStatus, nextURI *string, changed bool, err error) {
	/**
	At this point state and event are both non-nil and valid.

	State can be any one of the following:
	{status=enabled,  URI="foo"}
	{status=disabled, URI="foo"}
	{status=disabled, URI=""}

	Event can be any one of the following:
	{status=enabled,  URI="foo", defaultURI="bar"}
	{status=enabled,  URI="",    defaultURI="bar"}
	{status=disabled, URI="foo", defaultURI="bar"}
	{status=disabled, URI="",    defaultURI="bar"}
	{status=nil,      URI="foo", defaultURI="bar"}
	{status=nil,      URI="",    defaultURI="bar"}
	*/

	stateURISet := len(stateURI) != 0
	eventURISet := len(eventURI) != 0

	// factor this case out to ensure that URI is immutable
	if stateURISet && eventURISet && stateURI != eventURI {
		return nil, nil, false, errURIUpdate
	}

	// state 1
	if stateStatus == shared.ArchivalStatusEnabled && stateURISet {
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusDisabled), &stateURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && !eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusDisabled), &stateURI, true, nil
		}
		if eventStatus == nil && eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus == nil && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
	}

	// state 2
	if stateStatus == shared.ArchivalStatusDisabled && stateURISet {
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusEnabled), &stateURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && !eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusEnabled), &stateURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus == nil && eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus == nil && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
	}

	// state 3
	if stateStatus == shared.ArchivalStatusDisabled && !stateURISet {
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusEnabled), &eventURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusEnabled && !eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusEnabled), &eventDefaultURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusDisabled), &eventURI, true, nil
		}
		if eventStatus != nil && *eventStatus == shared.ArchivalStatusDisabled && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
		if eventStatus == nil && eventURISet {
			return common.ArchivalStatusPtr(shared.ArchivalStatusDisabled), &eventURI, true, nil
		}
		if eventStatus == nil && !eventURISet {
			return &stateStatus, &stateURI, false, nil
		}
	}
	return nil, nil, false, errCannotHandleStateChange
}
