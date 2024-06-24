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

package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/mapq/types"
)

type Dispatcher struct {
	consumer  types.Consumer
	ctx       context.Context
	cancelCtx context.CancelFunc
	wg        sync.WaitGroup
}

func New(c types.Consumer) *Dispatcher {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &Dispatcher{
		consumer:  c,
		ctx:       ctx,
		cancelCtx: cancelCtx,
	}
}

func (d *Dispatcher) Start(ctx context.Context) error {
	d.wg.Add(1)
	go d.run()
	return nil
}

func (d *Dispatcher) Stop(ctx context.Context) error {
	d.cancelCtx()
	timeout := 10 * time.Second
	if dl, ok := ctx.Deadline(); ok {
		timeout = time.Until(dl)
	}
	if !common.AwaitWaitGroup(&d.wg, timeout) {
		return fmt.Errorf("failed to stop dispatcher in %v", timeout)
	}
	return nil
}

func (d *Dispatcher) run() {
	defer d.wg.Done()
	// TODO: implement
}
