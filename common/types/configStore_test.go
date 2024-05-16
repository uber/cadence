package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicConfigEntryCopy(t *testing.T) {
	tests := []struct {
		name  string
		input *DynamicConfigEntry
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: &DynamicConfigEntry{},
		},
		{
			name: "nil values",
			input: &DynamicConfigEntry{
				Name:   "test-2",
				Values: nil,
			},
		},
		{
			name: "non-nil values",
			input: &DynamicConfigEntry{
				Name: "test-2",
				Values: []*DynamicConfigValue{
					&DynamicConfigValue{
						Value:   &DataBlob{Data: []byte("data")},
						Filters: nil,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input, tc.input.Copy())
		})
	}
}

func TestDynamicConfigFilterCopy(t *testing.T) {
	tests := []struct {
		name  string
		input *DynamicConfigFilter
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: &DynamicConfigFilter{},
		},
		{
			name: "nil value",
			input: &DynamicConfigFilter{
				Name:  "test-2",
				Value: nil,
			},
		},
		{
			name: "non-nil values",
			input: &DynamicConfigFilter{
				Name:  "test-2",
				Value: &DataBlob{Data: []byte("data")},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input, tc.input.Copy())
		})
	}
}
func TestDynamicConfigValueCopy(t *testing.T) {
	tests := []struct {
		name  string
		input *DynamicConfigValue
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: &DynamicConfigValue{},
		},

		{
			name: "non-nil values",
			input: &DynamicConfigValue{
				Value: &DataBlob{Data: []byte("data")},
				Filters: []*DynamicConfigFilter{
					&DynamicConfigFilter{
						Name:  "filter-1",
						Value: &DataBlob{Data: []byte("data")},
					}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input, tc.input.Copy())
		})
	}
}
