// Copyright (c) 2023 Uber Technologies, Inc.
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

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
)

func TestArchivalStateGetNextState(t *testing.T) {
	enabled := types.ArchivalStatusEnabled
	disabled := types.ArchivalStatusDisabled
	// Mock URI validation function that always succeeds
	successfulValidation := func(URI string) error { return nil }

	tests := []struct {
		name            string
		currentState    *ArchivalState
		event           *ArchivalEvent
		validateURI     func(string) error
		expectedState   *ArchivalState
		expectedChanged bool
		expectedError   error
	}{
		{
			name: "Enable archival with valid URI, no change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				URI:        "http://existing-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://existing-uri.com"},
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "Enable archival without setting a URI - triggers validation error",
			currentState: &ArchivalState{
				Status: disabled,
				URI:    "",
			},
			event: &ArchivalEvent{
				defaultURI: "",
				URI:        "",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errInvalidEvent, // Expecting error due to enabling archival without a URI
		},

		{
			name: "Archival disabled with URI, enabling with same URI",
			currentState: &ArchivalState{
				Status: disabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				URI:        "http://existing-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://existing-uri.com"},
			expectedChanged: true,
			expectedError:   nil,
		},
		{
			name: "Archival disabled without URI, enabling with default URI",
			currentState: &ArchivalState{
				Status: disabled,
				URI:    "",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://default-uri.com"},
			expectedChanged: true,
			expectedError:   nil,
		},
		{
			name: "Event with empty defaultURI",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errInvalidEvent,
		},
		{
			name: "Immutable URI change attempt",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				URI:        "http://new-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errURIUpdate,
		},
		{
			name:         "Current state is nil",
			currentState: nil,
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange,
		},
		{
			name: "Event is nil",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event:           nil,
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange,
		},
		{
			name:            "State and event nil",
			currentState:    nil,
			event:           nil,
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange,
		},
		{
			name: "Unrecognized ArchivalStatus in event",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     types.ArchivalStatus(999).Ptr(), // Assuming 999 is an unrecognized status
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange,
		},
		{
			name: "Starting from an invalid state configuration",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "", // Invalid state: enabled but no URI
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errInvalidState,
		},
		{
			name: "URI validation failure after passing immutability check",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				status:     &enabled,
				URI:        "http://existing-uri.com",
				defaultURI: "http://default-uri.com",
			},
			validateURI: func(URI string) error {
				// Simulating a validation failure for any URI, focusing on triggering validation logic inside defer
				return &types.BadRequestError{Message: "URI validation failed due to invalid format"}
			},
			expectedState:   nil,                                                                            // Expecting nil as the validation failure should prevent state transition
			expectedChanged: false,                                                                          // Expecting false as no change should occur due to validation failure
			expectedError:   &types.BadRequestError{Message: "URI validation failed due to invalid format"}, // Specific error from URI validation
		},
		{
			name: "Trigger defer logic without URI change",
			currentState: &ArchivalState{
				Status: types.ArchivalStatusEnabled,
				URI:    "http://existing-uri.com", // Keeping URI same to avoid errURIUpdate.
			},
			event: &ArchivalEvent{
				status:     &enabled,
				URI:        "http://existing-uri.com",
				defaultURI: "http://default-uri.com",
			},
			validateURI: func(URI string) error {
				return errCannotHandleStateChange
			},
			expectedState:   nil, // Adjusting expectation based on understanding of defer logic.
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange, // Expecting this validation failure.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextState, changed, err := tt.currentState.getNextState(tt.event, tt.validateURI)
			assert.Equal(t, tt.expectedState, nextState)
			assert.Equal(t, tt.expectedChanged, changed)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestNeverEnabledState(t *testing.T) {
	expectedState := &ArchivalState{
		URI:    "",
		Status: types.ArchivalStatusDisabled,
	}

	actualState := neverEnabledState()

	assert.Equal(t, expectedState, actualState, "The returned state does not match the expected never enabled state")
}

func TestArchivalStateGetNextState_State1Scenarios(t *testing.T) {
	enabled := types.ArchivalStatusEnabled
	disabled := types.ArchivalStatusDisabled
	// Mock URI validation function that always succeeds
	successfulValidation := func(URI string) error { return nil }

	tests := []struct {
		name            string
		currentState    *ArchivalState
		event           *ArchivalEvent
		validateURI     func(string) error
		expectedState   *ArchivalState
		expectedChanged bool
		expectedError   error
	}{
		{
			name: "Enabled status without event URI set - no change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     &enabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://existing-uri.com"},
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "Disabled status with event URI set - state change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				URI:        "http://existing-uri.com",
				status:     &disabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: disabled, URI: "http://existing-uri.com"},
			expectedChanged: true,
			expectedError:   nil,
		},
		{
			name: "Disabled status without event URI set - state change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				status:     &disabled,
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: disabled, URI: "http://existing-uri.com"},
			expectedChanged: true,
			expectedError:   nil,
		},
		{
			name: "No status with event URI set - no change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				URI:        "http://existing-uri.com", // URI set but no status change attempted.
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://existing-uri.com"},
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "No status without event URI set - no change",
			currentState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
				// No URI and no status change attempted.
			},
			validateURI:     successfulValidation,
			expectedState:   &ArchivalState{Status: enabled, URI: "http://existing-uri.com"},
			expectedChanged: false,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextState, changed, err := tt.currentState.getNextState(tt.event, tt.validateURI)
			assert.Equal(t, tt.expectedState, nextState)
			assert.Equal(t, tt.expectedChanged, changed)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestArchivalStateGetNextState_State2Scenarios(t *testing.T) {
	disabled := types.ArchivalStatusDisabled
	enabled := types.ArchivalStatusEnabled
	successfulValidation := func(URI string) error { return nil }

	currentStateWithURI := &ArchivalState{
		Status: disabled,
		URI:    "http://existing-uri.com",
	}

	tests := []struct {
		name            string
		event           *ArchivalEvent
		validateURI     func(string) error
		expectedState   *ArchivalState
		expectedChanged bool
		expectedError   error
	}{
		{
			name: "Enable archival without event URI - state change to enabled with existing URI",
			event: &ArchivalEvent{
				status:     &enabled,
				defaultURI: "http://default-uri.com",
			},
			validateURI: successfulValidation,
			expectedState: &ArchivalState{
				Status: enabled,
				URI:    "http://existing-uri.com",
			},
			expectedChanged: true,
			expectedError:   nil,
		},
		{
			name: "Disable archival with event URI set - no change",
			event: &ArchivalEvent{
				status:     &disabled,
				URI:        "http://existing-uri.com",
				defaultURI: "http://default-uri.com",
			},
			validateURI:     successfulValidation,
			expectedState:   currentStateWithURI,
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "Disable archival without event URI - no change",
			event: &ArchivalEvent{
				status:     &disabled,
				defaultURI: "http://default-uri.com",
			},
			validateURI:     successfulValidation,
			expectedState:   currentStateWithURI,
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "Event URI set without status change - no change",
			event: &ArchivalEvent{
				URI:        "http://existing-uri.com",
				defaultURI: "http://default-uri.com",
			},
			validateURI:     successfulValidation,
			expectedState:   currentStateWithURI,
			expectedChanged: false,
			expectedError:   nil,
		},
		{
			name: "No event URI or status change - no change",
			event: &ArchivalEvent{
				defaultURI: "http://default-uri.com",
			},
			validateURI:     successfulValidation,
			expectedState:   currentStateWithURI,
			expectedChanged: false,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextState, changed, err := currentStateWithURI.getNextState(tt.event, tt.validateURI)
			assert.Equal(t, tt.expectedState, nextState)
			assert.Equal(t, tt.expectedChanged, changed)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
