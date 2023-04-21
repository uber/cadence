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

package persistence

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/types"
)

func TestClusterReplicationConfigGetCopy(t *testing.T) {
	config := &ClusterReplicationConfig{
		ClusterName: "test",
	}
	assert.Equal(t, config, config.GetCopy()) // deep equal
	assert.Equal(t, true, config != config.GetCopy())
}

func TestIsTransientError(t *testing.T) {
	transientErrors := []error{
		&types.ServiceBusyError{},
		&types.InternalServiceError{},
		&TimeoutError{},
	}
	for _, err := range transientErrors {
		require.True(t, IsTransientError(err))
	}

	nonRetryableErrors := []error{
		&types.EntityNotExistsError{},
		&types.DomainAlreadyExistsError{},
		&WorkflowExecutionAlreadyStartedError{},
		errors.New("some unknown error"),
	}
	for _, err := range nonRetryableErrors {
		require.False(t, IsTransientError(err))
	}
}

func TestIsolationGroupSerialization(t *testing.T) {

	tests := map[string]struct {
		in          []types.IsolationGroupPartition
		expectedErr error
	}{
		"valid case": {
			in: []types.IsolationGroupPartition{
				{Name: "zone-1", State: types.IsolationGroupStateDrained},
				{Name: "zone-2", State: types.IsolationGroupStateHealthy},
			},
		},
		"empty case": {
			in: []types.IsolationGroupPartition{},
		},
		"nil case": {
			in: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := DeserializeIsolationGroups(SerializeIsolationGroups(td.in))
			assert.Equal(t, td.in, out)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}
func TestIsolationGroupDeserialization(t *testing.T) {

	tests := map[string]struct {
		in          []map[string]interface{}
		expectedOut []types.IsolationGroupPartition
		expectedErr error
	}{
		"valid case": {
			in: []map[string]interface{}{
				{
					"name":  "zone-1",
					"state": 2,
				},
				{
					"name":  "zone-2",
					"state": 1,
				},
			},
			expectedOut: []types.IsolationGroupPartition{
				{Name: "zone-1", State: types.IsolationGroupStateDrained},
				{Name: "zone-2", State: types.IsolationGroupStateHealthy},
			},
		},
		"nil case": {
			in:          nil,
			expectedOut: nil,
		},
		"empty case": {
			in:          []map[string]interface{}{},
			expectedOut: []types.IsolationGroupPartition{},
		},
		"invalid state case 1": {
			in: []map[string]interface{}{
				{
					"name":  "bad value",
					"state": "wrong type",
				},
			},
			expectedErr: errors.New("failed to get correct type while deserializing isolation groups: [map[name:bad value state:wrong type]], row: map[name:bad value state:wrong type]"),
		},
		"invalid state case 2": {
			in: []map[string]interface{}{
				{
					"name": "missing value",
				},
			},
			expectedErr: errors.New("failed to deserialize isolation groups: [map[name:missing value]], row: map[name:missing value]"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := DeserializeIsolationGroups(td.in)
			assert.Equal(t, td.expectedOut, out)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}
