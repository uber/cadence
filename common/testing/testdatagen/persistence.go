package testdatagen

import (
	"github.com/google/gofuzz"
	"github.com/uber/cadence/common/types"
)

// GenHistoryEvent is a function to use with gofuzz which
// skips the majority of difficult to generate values
// for the sake of simplicity in testing. Use it with the fuzz.Funcs(...) generation function
func GenHistoryEvent(o *types.HistoryEvent, c fuzz.Continue) {
	t1 := int64(1704127933)
	e1 := types.EventType(1)
	*o = types.HistoryEvent{
		ID:        1,
		Timestamp: &t1,
		EventType: &e1,
		Version:   2,
		TaskID:    3,
		// todo (david.porter) Find a better way to handle substructs
	}
	return
}
