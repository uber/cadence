package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsolationGroupConfiguration_ToPartitionList(t *testing.T) {
	tests := map[string]struct {
		in          *IsolationGroupConfiguration
		expectedOut []IsolationGroupPartition
	}{
		"valid value": {
			in: &IsolationGroupConfiguration{
				"zone-1": {
					Name:  "zone-1",
					State: IsolationGroupStateDrained,
				},
				"zone-2": {
					Name:  "zone-2",
					State: IsolationGroupStateHealthy,
				},
				"zone-3": {
					Name:  "zone-3",
					State: IsolationGroupStateDrained,
				},
			},
			expectedOut: []IsolationGroupPartition{
				{
					Name:  "zone-1",
					State: IsolationGroupStateDrained,
				},
				{
					Name:  "zone-2",
					State: IsolationGroupStateHealthy,
				},
				{
					Name:  "zone-3",
					State: IsolationGroupStateDrained,
				},
			},
		},
	}
	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedOut, td.in.ToPartitionList())
		})
	}
}
