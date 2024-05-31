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

package collection

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/exp/slices"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
)

func TestIt(t *testing.T) {
	ctrl := gomock.NewController(t)
	l, obs := testlogger.NewObserved(t)
	c := New(quotas.NewMockLimiterFactory(ctrl), func(opts ...dynamicconfig.FilterOption) time.Duration {
		return time.Second
	}, l, metrics.NewNoopMetricsClient().Scope(metrics.GetGlobalIsolationGroups)) // TODO: make scope
	ts := clock.NewMockedTimeSource()
	c.TestOverrides(t, ts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, c.OnStart(ctx))
	ts.BlockUntil(1)        // created timer in background updater
	ts.Advance(time.Second) // trigger the update
	// TODO: how do we assert "consumed from channel and is processing it"?  is that even possible?
	time.Sleep(time.Millisecond)
	require.NoError(t, c.OnStop(ctx))
	// will fail if the tick did not begin processing, e.g. no tick or it raced with shutdown.
	// without a brief sleep after advancing time, this extremely rarely fails (try 10k tests to see)
	assert.Len(t, obs.FilterMessage("update tick").All(), 1, "no update logs, did it process the tick? all: %#v", obs.All())

	goleak.VerifyNone(t) // must be shut down
}

func TestObserve(t *testing.T) {
	/*
		tickers can:
		- get Chan()
		- buffer one event
		- reset to drain and change timing
	*/
	buf := make(chan int, 2)
	unbuf := make(chan int)
	stop := make(chan int)

	var done sync.WaitGroup

	// CHECK: can I read-and-write in a case and assert "could proxy"?
	// RESULT: nope.  reads from buf before entering select, then waits on pushing to unbuf.
	// does not read-and-write in one op.
	//
	//
	// done.Add(1)
	// go func() {
	// 	defer done.Done()
	// 	for {
	// 		fmt.Println("selecting")
	// 		select {
	// 		case unbuf <- <-buf:
	// 			fmt.Println("pushed to unbuf")
	// 		case <-stop:
	// 			fmt.Println("stopping")
	// 			return
	// 		}
	// 	}
	// }()
	// done.Add(1)
	// go func() {
	// 	defer done.Done()
	// 	select {
	// 	case v := <-unbuf:
	// 		fmt.Println("unbuf val:", v)
	// 	case <-stop:
	// 		return
	// 	}
	// }()
	//
	// // buf <- 1
	// time.Sleep(time.Millisecond)
	// fmt.Println("stop")

	// CHECK: on an unbuffered chan vs a closed chan, does a successful write always mean "consumed" not "raced with closed and dropped"?
	done.Add(1)

	go func() {
		defer done.Done()
		select {
		case v := <-unbuf:
			buf <- v
		case <-stop:
			buf <- 0
		}
	}()
	done.Add(1)
	go func() {
		defer done.Done()
		close(stop)
	}()
	done.Add(1)
	go func() {
		defer done.Done()
		select {
		case unbuf <- 1:
			buf <- 2
		case <-time.After(time.Microsecond):
			buf <- 3
		}
	}()

	ops := []int{<-buf, <-buf}
	slices.Sort(ops)
	if ops[0] == 0 && ops[1] == 2 {
		t.Error("closed and wrote")
	} else if ops[0] == 1 && ops[1] == 3 {
		t.Error("timeout and read, should be impossible")
	}

	done.Wait()
}
