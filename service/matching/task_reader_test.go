// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package matching

import (
	"testing"

	"go.uber.org/goleak"

	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
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
