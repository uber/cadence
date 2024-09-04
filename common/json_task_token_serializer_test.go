package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJsonTaskTokenSerializer_Token_RoundTrip(t *testing.T) {
	token := TaskToken{
		DomainID:        "test-domain",
		WorkflowID:      "test-workflow-id",
		WorkflowType:    "test-workflow-type",
		RunID:           "test-run-id",
		ScheduleID:      1,
		ScheduleAttempt: 1,
		ActivityID:      "test-activity-id",
		ActivityType:    "test-activity-type",
	}

	serializer := NewJSONTaskTokenSerializer()

	serialized, err := serializer.Serialize(&token)
	require.NoError(t, err)
	deserializedToken, err := serializer.Deserialize(serialized)
	require.NoError(t, err)
	require.Equal(t, token, *deserializedToken)
}

func TestNewJSONTaskTokenSerializer_QueryToken_Roundtrip(t *testing.T) {
	token := QueryTaskToken{
		DomainID:     "test-domain",
		WorkflowID:   "test-workflow-id",
		WorkflowType: "test-workflow-type",
		RunID:        "test-run-id",
		TaskList:     "test-task-list",
		TaskID:       "test-task-id",
	}
	serializer := NewJSONTaskTokenSerializer()

	serialized, err := serializer.SerializeQueryTaskToken(&token)
	require.NoError(t, err)
	deserializedToken, err := serializer.DeserializeQueryTaskToken(serialized)
	require.NoError(t, err)
	require.Equal(t, token, *deserializedToken)
}
