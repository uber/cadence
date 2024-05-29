package testlogger

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/uber/cadence/common/log/tag"
)

var (
	done   = make(chan struct{})
	logged = make(chan struct{})
)

func TestMain(m *testing.M) {
	code := m.Run()
	// ensure synchronization between t.done and t.logf, else this test is extremely flaky.
	// for details see: https://github.com/golang/go/issues/67701
	close(done)
	select {
	case <-logged:
		os.Exit(code)
	case <-time.After(time.Second): // should be MUCH faster
		// also fails the test due to exit(1)
		log.Fatal("timed out waiting for test")
	}
}

// Unfortunately a moderate hack, to work around our faulty lifecycle management,
// and some libraries with issues as well.
// Ideally this test WOULD fail, but that's much harder to assert "safely".
func TestLoggerShouldNotFailIfLoggedLate(t *testing.T) {
	origLogger := New(t)
	// if With does not defer core selection, this will fail the test
	// by sending the logs to t.Logf
	withLogger := origLogger.WithTags(tag.ActorID("testing"))
	go func() {
		<-done
		origLogger.Info("too late, orig")
		withLogger.Info("too late, with")
		close(logged)
	}()
}
