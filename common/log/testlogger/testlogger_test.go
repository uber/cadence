package testlogger

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/uber/cadence/common/log"
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
		_, _ = fmt.Fprintln(os.Stderr, "timed out waiting for test to log")
		os.Exit(1)
	}
}

// Unfortunately a moderate hack, to work around our faulty lifecycle management,
// and some libraries with issues as well.
// Ideally this test WOULD fail, but that's much harder to assert "safely".
func TestLoggerShouldNotFailIfLoggedLate(t *testing.T) {
	origLogger := New(t)
	// if With does not defer core selection, this will fail the test
	// by sending the logs to t.Logf
	withLogger := origLogger.WithTags(tag.ActorID("testing")) // literally any tag
	origLogger.Info("before is fine, orig")
	withLogger.Info("before is fine, with")
	go func() {
		<-done
		origLogger.Info("too late, orig")
		withLogger.Info("too late, with")
		close(logged)
	}()
}

func TestSubtestShouldNotFail(t *testing.T) {
	// when complete, a subtest's too-late logs just get pushed to the parent,
	// and do not fail any tests.  they only fail when no running parent exists.
	//
	// if Go changes this behavior, this test could fail, otherwise AFAICT it
	// should be stable.
	assertDoesNotFail := func(name string, setup, log func(t *testing.T)) {
		// need to wrap in something that will out-live the "real" test,
		// to ensure there is a running parent test to push logs toward.
		t.Run(name, func(t *testing.T) {
			// same setup as TestMain but contained within this sub-test
			var (
				done   = make(chan struct{})
				logged = make(chan struct{})
			)
			t.Run("inner", func(t *testing.T) {
				setup(t)
				go func() {
					<-done
					// despite being too late, the parent test is still running
					// so this does not fail the test.
					log(t)
					close(logged)
				}()
				time.AfterFunc(10*time.Millisecond, func() {
					close(done)
				})
			})
			<-logged
		})
	}

	assertDoesNotFail("real", func(t *testing.T) {
		// no setup needed
	}, func(t *testing.T) {
		t.Logf("too late but allowed")
	})

	var l log.Logger
	assertDoesNotFail("wrapped", func(t *testing.T) {
		l = New(t)
	}, func(t *testing.T) {
		l.Info("too late but allowed")
	})
}
