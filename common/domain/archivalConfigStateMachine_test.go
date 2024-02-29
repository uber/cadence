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
