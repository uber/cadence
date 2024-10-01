package isolationgroupapi

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapAllIsolationGroupStates(t *testing.T) {

	tests := map[string]struct {
		in          []interface{}
		expected    []string
		expectedErr error
	}{
		"valid mapping": {
			in:       []interface{}{"zone-1", "zone-2", "zone-3"},
			expected: []string{"zone-1", "zone-2", "zone-3"},
		},
		"invalid mapping": {
			in:          []interface{}{1, 2, 3},
			expectedErr: errors.New("failed to get all-isolation-groups response from dynamic config: got 1 (int)"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := MapAllIsolationGroupsResponse(td.in)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}
