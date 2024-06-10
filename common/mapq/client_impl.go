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

package mapq

import (
	"context"

	"github.com/uber/cadence/common/mapq/tree"
	"github.com/uber/cadence/common/mapq/types"
)

type ClientImpl struct {
	persister       types.Persister
	consumerFactory types.ConsumerFactory
	tree            *tree.QueueTree
	partitions      []string
	policies        []types.NodePolicy
}

func (c *ClientImpl) Start(ctx context.Context) error {
	return c.tree.Start(ctx)
}

func (c *ClientImpl) Stop(ctx context.Context) error {
	// Stop the tree which will stop the dispatchers
	if err := c.tree.Stop(ctx); err != nil {
		return err
	}

	// stop the consumer factory which will stop the consumers
	return c.consumerFactory.Stop(ctx)
}

func (c *ClientImpl) Enqueue(ctx context.Context, items []types.Item) error {
	return c.tree.Enqueue(ctx, items)
}

func (c *ClientImpl) Ack(context.Context, types.Item) error {
	// TODO: implement
	return nil
}

func (c *ClientImpl) Nack(context.Context, types.Item) error {
	// TODO: implement
	return nil
}
