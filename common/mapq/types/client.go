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

import (
	"context"
)

type Client interface {
	// Enqueue adds an item to the queue
	Enqueue(context.Context, []Item) ([]ItemToPersist, error)

	// Ack marks the item as processed.
	// Out of order acks are supported.
	// Acking an item that has not been dequeued will have no effect.
	// Queue's committed offset is updated to the last ack'ed item's offset periodically until there's a gap (un-acked item).
	// In other words, all items up to the committed offset are ack'ed.
	// Ack'ed item might still be returned from Dequeue if its offset is higher than last committed offset before process restarts.
	Ack(context.Context, Item) error

	// Nack negatively acknowledges an item in the queue
	// Nack'ing an already ack'ed item will have no effect.
	Nack(context.Context, Item) error

	// Start the client. It will
	// - fetch the last committed offsets from the persister,
	// - start corresponding consumers
	// - dispatch items starting from those offsets.
	Start(context.Context) error

	// Stop the client. It will
	// - stop all consumers
	// - persist the last committed offsets
	Stop(context.Context) error
}
