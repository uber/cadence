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

package migration

import (
	"context"
	"fmt"
	"sync"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

const (
	ReaderRolloutCallOnlyOldAndReturnOld ReaderRolloutState = "only-call-old"
	ReaderRolloutCallBothAndReturnOld    ReaderRolloutState = "call-both-return-old"
	ReaderRolloutCallBothAndReturnNew    ReaderRolloutState = "call-both-return-new"
	ReaderRolloutCallOnlyNewAndReturnNew ReaderRolloutState = "only-call-new"
)

type resultPair[T any] struct {
	res T
	err error
}

func (c *readerImpl[T]) ReadAndReturnActive(
	ctx context.Context,
	constraints Constraints,
	old func(ctx context.Context) (*T, error),
	new func(ctx context.Context) (*T, error),
) (*T, error) {

	state := c.getReaderRolloutState(constraints)

	var active func(ctx context.Context) (*T, error)
	var background func(ctx context.Context) (*T, error)

	switch state {
	case ReaderRolloutCallOnlyOldAndReturnOld:
		return old(ctx)
	case ReaderRolloutCallBothAndReturnOld:
		active = old
		background = new
	case ReaderRolloutCallBothAndReturnNew:
		active = new
		background = old
	case ReaderRolloutCallOnlyNewAndReturnNew:
		return new(ctx)
	default:
		return old(ctx)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	// 0 is always active data, index 1 is background
	// using a slice to avoid locking
	results := make([]resultPair[*T], 2)

	// run the shadow request in a goroutine, wait for the results and then gather them up
	go func() {
		defer func() {
			log.CapturePanic(recover(), c.log, nil)
			c.scope.IncCounter(metrics.ShadowerBackgroundPanicRecoverCount)
		}()
		defer wg.Done()

		backgroundRes, backgroundErr := c.callBackgroundAndCompare(ctx, background)
		results[1].res = backgroundRes
		results[1].err = backgroundErr
	}()

	// Fire off a watcher which is waiting for both sets of results to come through, and then once done so
	// it'll hand them to the comparison function for analysis
	go func() {
		if !c.shouldCompare(constraints) {
			return
		}
		defer func() {
			log.CapturePanic(recover(), c.log, nil)
			c.scope.IncCounter(metrics.ShadowerBackgroundPanicRecoverCount)
		}()

		wg.Wait()
		c.comparisonFn(c.log, c.scope, results[0].res, results[0].err, results[1].res, results[1].err)
	}()

	activeRes, activeErr := active(ctx)
	results[0].res = activeRes
	results[0].err = activeErr
	wg.Done()
	return activeRes, activeErr
}

// Fires off a request on a background goroutine and waits for it to complete OR when the hard deadline
// in backgroundTimeout expires. The op() code is expected to be faulty, to possibly panic, take too long
// or otherwise misbehave and so the hard deadline is intended as a last-ditch safety measure if the context
// is not being obeyed
func (c *readerImpl[T]) callBackgroundAndCompare(parentCtx context.Context, op func(ctx context.Context) (*T, error)) (*T, error) {

	newRequestCtx, cancel := copyParentContextWhileBreakingParentCancellation(parentCtx)
	defer cancel()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), c.backgroundTimeout)
	defer cancel()

	completion := make(chan resultPair[*T])
	go func() {
		defer func() {
			log.CapturePanic(recover(), c.log, nil)
			c.scope.IncCounter(metrics.ShadowerBackgroundPanicRecoverCount)
		}()
		defer close(completion)

		res, err := op(newRequestCtx)
		completion <- resultPair[*T]{
			res: res,
			err: err,
		}
	}()

	// now return whatever finishes first
	select {
	case <-timeoutCtx.Done():
		c.scope.IncCounter(metrics.ShadowerBackgroundRequestHardCutCount)
		return nil, fmt.Errorf("shadower cut off a request taking too long: %w", timeoutCtx.Err())
	case resPair := <-completion:
		return resPair.res, resPair.err
	}
}

// the intention here is to allow this child context to continue running until completion
// because in shadowing, you're necessarily not going to want to block on a slow/maybe not working
// new codepath, but we do need to wait for results for both (to compare). So when the active flow
// is finishing first, we want to not have it's cancellation kill off the background context.
func copyParentContextWhileBreakingParentCancellation(ctx context.Context) (c context.Context, cancelFunc func()) {
	parentDeadline, ok := ctx.Deadline()
	if !ok {
		return ctx, func() {}
	}

	return context.WithDeadline(context.WithoutCancel(ctx), parentDeadline)
}
