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
	"fmt"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/mapq/tree"
	"github.com/uber/cadence/common/mapq/types"
	"github.com/uber/cadence/common/metrics"
)

type Options func(*clientImpl)

func WithPersister(p types.Persister) Options {
	return func(c *clientImpl) {
		c.persister = p
	}
}

func WithConsumerFactory(cf types.ConsumerFactory) Options {
	return func(c *clientImpl) {
		c.consumerFactory = cf
	}
}

// WithPartitions sets the partition keys for each level.
// MAPQ creates a tree with depth = len(partitions)
func WithPartitions(partitions []string) Options {
	return func(c *clientImpl) {
		c.partitions = partitions
	}
}

// WithPolicies sets the policies for the MAPQ instance.
// Policies can be defined for nodes at a specific level or nodes with specific path.
//
// Path conventions:
// - "*" -> represents the root node at level 0
// - "*/." matches with all nodes at level 1
// - "*/*" represents the catch-all node at level 1
// - "*/xyz" represents a specific node at level 1 whose partition value is xyz
// - "*/./." matches with all nodes at level 2
// - "*/xyz/." matches with all nodes at level 2 whose parent is xyz node
// - "*/xyz/*" represents the catch-all node at level 2 whose parent is xyz node
// - "*/xyz/abc" represents a specific node at level 2 whose level 2 attribute value is abc and parent is xyz node
func WithPolicies(policies []types.NodePolicy) Options {
	return func(c *clientImpl) {
		c.policies = policies
	}
}

func New(logger log.Logger, scope metrics.Scope, opts ...Options) (types.Client, error) {
	c := &clientImpl{
		logger: logger.WithTags(tag.ComponentMapQ),
		scope:  scope,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.persister == nil {
		return nil, fmt.Errorf("persister is required. Use WithPersister option to set it")
	}

	if c.consumerFactory == nil {
		return nil, fmt.Errorf("consumer factory is required. Use WithConsumerFactory option to set it")
	}

	tree, err := tree.New(logger, scope, c.partitions, c.policies, c.persister, c.consumerFactory)
	if err != nil {
		return nil, err
	}

	c.tree = tree
	c.logger.Info("MAPQ client created",
		tag.Dynamic("partitions", c.partitions),
		tag.Dynamic("policies", c.policies),
		tag.Dynamic("tree", c.tree.String()),
	)

	return c, nil
}
