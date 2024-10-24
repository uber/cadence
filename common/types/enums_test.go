package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_EventTypeValues(t *testing.T) {
	require.ElementsMatch(t, _eventTypeValues, EventTypeValues())
}

func Test_DecisionTypeValues(t *testing.T) {
	require.ElementsMatch(t, _decisionTypeValues, DecisionTypeValues())
}
