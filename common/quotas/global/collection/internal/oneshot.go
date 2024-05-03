package internal

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"go.uber.org/multierr"
)

// OneShot simplifies one-time startup and shutdown logic:
//   - Start can be called once with an init func which will be run (duplicate calls will error and do nothing)
//   - Stop can be called once, after which it will error (but stopping will continue)
//   - Stopped can be called only once or it will panic (as errors are typically not reasonable to return to a caller from this action).
//     Always call it in only one location that can only be triggered by Stop.
//   - WaitForShutdown can be called any number of times, and will wait until Stopped is called (returns nil) or the context times out (returns ctx.Err())
//
// This makes it essentially a validating sync.Once with wait-able completion.
type OneShot struct {
	started atomic.Bool
	stop    atomic.Bool

	ctx     context.Context
	cancel  context.CancelFunc
	stopped chan struct{}
}

func NewOneShot() *OneShot {
	ctx, cancel := context.WithCancel(context.Background())
	return &OneShot{
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}
}

func (o *OneShot) Start(do func() error) error {
	if !o.started.CompareAndSwap(false, true) {
		return errors.New("*OneShot was already started")
	}
	err := do()
	if err != nil {
		return fmt.Errorf("*OneShot failed to start: %w", err)
	}
	return nil
}

func (o *OneShot) Stop() (err error) {
	if !o.stop.CompareAndSwap(false, true) {
		return errors.New("*OneShot was already stopped")
	}
	o.cancel()
	return
}

func (o *OneShot) Context() context.Context {
	return o.ctx
}

func (o *OneShot) StopAndWait(ctx context.Context) error {
	stopErr := o.Stop()               // duplicate calls are wrong...
	waitErr := o.WaitForShutdown(ctx) // ... but wait either way.
	combined := multierr.Append(stopErr, waitErr)
	return combined
}

func (o *OneShot) Stopped() {
	close(o.stopped) // allowed to panic if called more than once - this should never be used multiple times
}

func (o *OneShot) WaitForShutdown(ctx context.Context) error {
	select {
	case <-o.stopped:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("*OneShot shutdown timed out: %w", ctx.Err())
	}
}
