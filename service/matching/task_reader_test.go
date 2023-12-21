package matching

import (
	"testing"

	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
	"go.uber.org/goleak"
)

func TestMultipleStops(t *testing.T) {
	// Make sure that taskReader is fully stopped.
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	logger := testlogger.New(t)
	tl := &taskReader{
		logger:         logger,
		taskAckManager: messaging.NewAckManager(logger),
		taskGC: &taskGC{
			config: &taskListConfig{
				MaxTaskDeleteBatchSize: func() int { return 0 },
			},
		},
	}
	waitCh := make(chan struct{})
	tl.cancelFunc = func() {
		<-waitCh
	}
	go tl.Stop()

	stoppedCh := make(chan struct{})
	go func() {
		tl.Stop()
		close(stoppedCh)
	}()
	select {
	case <-stoppedCh:
		t.Error("task reader should not be stopped")
		t.Fail()
	case waitCh <- struct{}{}:
		close(waitCh)
	}
}
