package matching

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTaskInfoGetters(t *testing.T) {
	assert.NotPanics(t, func() {
		t := &InternalTask{}

		t.GetTaskID()
		t.GetDomainID()
		t.GetWorkflowID()
		t.GetWorkflowScheduleID()
		t.GetRunID()
	})
}
