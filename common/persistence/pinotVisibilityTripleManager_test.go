package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterAttrPrefix(t *testing.T) {
	tests := map[string]struct {
		expectedInput  string
		expectedOutput string
	}{
		"Case1: empty input": {
			expectedInput:  "",
			expectedOutput: "",
		},
		"Case2: filtered input": {
			expectedInput:  "`Attr.CustomIntField` = 12",
			expectedOutput: "CustomIntField = 12",
		},
		"Case3: complex input": {
			expectedInput:  "WorkflowID = 'test-wf' and (`Attr.CustomIntField` = 12 or `Attr.CustomStringField` = 'a-b-c' and WorkflowType = 'wf-type')",
			expectedOutput: "WorkflowID = 'test-wf' and (CustomIntField = 12 or CustomStringField = 'a-b-c' and WorkflowType = 'wf-type')",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualOutput := filterAttrPrefix(test.expectedInput)
				assert.Equal(t, test.expectedOutput, actualOutput)
			})
		})
	}
}
