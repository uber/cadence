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

package types

import "context"

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination consumer_mock.go -package types github.com/uber/cadence/common/mapq/types ConsumerFactory,Consumer

type ConsumerFactory interface {
	// New creates a new consumer with the given partitions or returns an existing consumer
	// to process the given partitions
	// Consumer lifecycle is managed by the factory so the returned consumer must be started.
	New(ItemPartitions) (Consumer, error)

	// Stop stops all consumers created by this factory
	Stop(context.Context) error
}

type Consumer interface {
	Start(context.Context) error
	Stop(context.Context) error
	Process(context.Context, Item) error
}
