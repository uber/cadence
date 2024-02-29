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
			name:            "State and event nil",
			currentState:    nil,
			event:           nil,
			validateURI:     successfulValidation,
			expectedState:   nil,
			expectedChanged: false,
			expectedError:   errCannotHandleStateChange,
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
