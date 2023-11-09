package taskvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/log"
)

func TestWorkflowCheckforValidation(t *testing.T) {
	// Create a logger for testing
	logger := log.NewNoop()

	// Create a new Checker
	checker := NewWfChecker(logger)

	// Test with sample data
	err := checker.WorkflowCheckforValidation("workflow123", "domain456", "run789")

	// Assert that there is no error
	assert.Nil(t, err)
}
