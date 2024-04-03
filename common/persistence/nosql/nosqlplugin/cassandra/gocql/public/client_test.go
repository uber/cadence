package public

import (
	"context"
	"fmt"
	gogocql "github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockError to simulate gocql.Error behavior
type MockError struct {
	gogocql.RequestError
	code    int
	message string
}

func (m MockError) Code() int {
	return m.code
}

func (m MockError) Message() string {
	return m.message
}

func TestClient_IsTimeoutError(t *testing.T) {
	client := client{}
	errorMap := map[error]bool{
		nil:                               false,
		context.DeadlineExceeded:          true,
		gogocql.ErrTimeoutNoResponse:      true,
		gogocql.ErrConnectionClosed:       true,
		&gogocql.RequestErrWriteTimeout{}: true,
		gogocql.ErrFrameTooBig:            false,
	}
	for err, expected := range errorMap {
		assert.Equal(t, expected, client.IsTimeoutError(err))
	}
}

func TestClient_IsNotFoundError(t *testing.T) {
	client := client{}
	errorMap := map[error]bool{
		nil:                    false,
		gogocql.ErrNotFound:    true,
		gogocql.ErrFrameTooBig: false,
	}
	for err, expected := range errorMap {
		assert.Equal(t, expected, client.IsNotFoundError(err))
	}
}

// TestClient_IsThrottlingError tests the IsThrottlingError function with different error codes
func TestClient_IsThrottlingError(t *testing.T) {
	client := client{}
	tests := []struct {
		name               string
		mockErrorCode      int
		expectedResult     bool
		nonCompatibleError error
	}{
		{
			name:           "With Throttling Error",
			mockErrorCode:  0x1001,
			expectedResult: true,
		},
		{
			name:               "With Non-Throttling Error",
			mockErrorCode:      0x0001,
			expectedResult:     false,
			nonCompatibleError: fmt.Errorf("with Non-Throttling Error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nonCompatibleError != nil {
				result := client.IsThrottlingError(tt.nonCompatibleError)
				assert.False(t, result)
			}
			err := MockError{code: tt.mockErrorCode}
			result := client.IsThrottlingError(err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestClient_IsDBUnavailableError(t *testing.T) {
	client := client{}
	tests := []struct {
		name               string
		mockMessage        string
		mockErrorCode      int
		expectedResult     bool
		nonCompatibleError error
	}{
		{
			name:           "With DB Unavailable Error",
			mockMessage:    "Cannot perform LWT operation",
			mockErrorCode:  0x1000,
			expectedResult: true,
		},
		{
			name:           "With Non-DB Unavailable Error",
			mockMessage:    "Cannot perform LWT operation",
			mockErrorCode:  0x0001,
			expectedResult: false,
		},
		{
			name:               "With Non-compatible Error",
			mockErrorCode:      0x0001,
			expectedResult:     false,
			nonCompatibleError: fmt.Errorf("with Non-compatible Error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nonCompatibleError != nil {
				result := client.IsDBUnavailableError(tt.nonCompatibleError)
				assert.False(t, result)
			}
			err := MockError{code: tt.mockErrorCode, message: tt.mockMessage}
			result := client.IsDBUnavailableError(err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestClient_IsCassandraConsistencyError(t *testing.T) {
	client := client{}
	tests := []struct {
		name               string
		mockErrorCode      int
		expectedResult     bool
		nonCompatibleError error
	}{
		{
			name:           "With Cassandra Consistency Error",
			mockErrorCode:  0x1000,
			expectedResult: true,
		},
		{
			name:           "With Non-Cassandra Consistency Error",
			mockErrorCode:  0x0001,
			expectedResult: false,
		},
		{
			name:               "With Non-compatible Error",
			mockErrorCode:      0x0001,
			expectedResult:     false,
			nonCompatibleError: fmt.Errorf("with Non-compatible Error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nonCompatibleError != nil {
				result := client.IsCassandraConsistencyError(tt.nonCompatibleError)
				assert.False(t, result)
			}
			err := MockError{code: tt.mockErrorCode}
			result := client.IsCassandraConsistencyError(err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
