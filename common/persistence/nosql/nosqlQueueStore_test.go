package nosql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNextID(t *testing.T) {
	tests := map[string]struct {
		acks     map[string]int64
		lastID   int64
		expected int64
	}{
		"expected case - last ID is equal to ack-levels": {
			acks:     map[string]int64{"a": 3},
			lastID:   3,
			expected: 4,
		},
		"expected case - last ID is equal to ack-levels haven't caught up": {
			acks:     map[string]int64{"a": 2},
			lastID:   3,
			expected: 4,
		},
		"error case - ack-levels are ahead for some reason": {
			acks:     map[string]int64{"a": 3},
			lastID:   2,
			expected: 4,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, getNextID(td.acks, td.lastID))
		})
	}
}
