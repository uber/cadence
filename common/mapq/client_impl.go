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
	"errors"
	"fmt"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mapq/tree"
	"github.com/uber/cadence/common/mapq/types"
	"github.com/uber/cadence/common/metrics"
)

type clientImpl struct {
	logger          log.Logger
	scope           metrics.Scope
	persister       types.Persister
	consumerFactory types.ConsumerFactory
	tree            *tree.QueueTree
	partitions      []string
	policies        []types.NodePolicy
}

func (c *clientImpl) Start(ctx context.Context) error {
	c.logger.Info("Starting MAPQ client")
	err := c.tree.Start(ctx)
	if err != nil {
		return err
	}

	c.logger.Info("Started MAPQ client")
	return nil
}

func (c *clientImpl) Stop(ctx context.Context) error {
	c.logger.Info("Stopping MAPQ client")

	// Stop the tree which will stop the dispatchers
	if err := c.tree.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop tree: %w", err)
	}

	// stop the consumer factory which will stop the consumers
	err := c.consumerFactory.Stop(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop consumer factory: %w", err)
	}

	c.logger.Info("Stopped MAPQ client")
	return nil
}

func (c *clientImpl) Enqueue(ctx context.Context, items []types.Item) ([]types.ItemToPersist, error) {
	return c.tree.Enqueue(ctx, items)
}

func (c *clientImpl) Ack(context.Context, types.Item) error {
	return errors.New("not implemented")
}

func (c *clientImpl) Nack(context.Context, types.Item) error {
	return errors.New("not implemented")
}
