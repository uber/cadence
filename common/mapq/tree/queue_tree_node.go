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

package tree

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/mapq/dispatcher"
	"github.com/uber/cadence/common/mapq/types"
	"github.com/uber/cadence/common/metrics"
)

// QueueTreeNode represents a node in the queue tree
type QueueTreeNode struct {
	// originalLogger is the logger passed in during creation. No node specific tags are added to this logger and it should be passed to child nodes
	originalLogger log.Logger

	// logger is the logger for this node. It has the node specific tags added to it
	logger log.Logger
	scope  metrics.Scope

	// The path to the node
	Path string

	// The partition key used by this node to decide which child to enqueue the item.
	// Partition key of a node is the attribute key of child node.
	PartitionKey string

	// The attribute key used to create this node by parent
	AttributeKey string

	// The attribute value used to create this node by parent
	AttributeVal any

	// The policy for this node. It's merged policy from all policies that match this node
	NodePolicy types.NodePolicy

	// Children by attribute key
	// "*" is a special key that represents the default/fallback child queue
	// If there's no children then the node is considered leaf node
	Children map[any]*QueueTreeNode

	// The dispatcher for this node. Only leaf nodes have dispatcher
	Dispatcher *dispatcher.Dispatcher
}

func (n *QueueTreeNode) Start(
	ctx context.Context,
	consumerFactory types.ConsumerFactory,
	partitions []string,
	partitionMap map[string]any,
) error {
	n.logger.Info("Starting node", tag.Dynamic("node", n.String()))

	// If there are no children then this is a leaf node
	if len(n.Children) == 0 {
		n.logger.Info("Creating consumer and starting a new dispatcher for leaf node")
		c, err := consumerFactory.New(types.NewItemPartitions(partitions, partitionMap))
		if err != nil {
			return err
		}
		d := dispatcher.New(c)
		if err := d.Start(ctx); err != nil {
			return err
		}
		n.Dispatcher = d
		return nil
	}

	for _, child := range n.Children {
		partitionMap[n.PartitionKey] = child.AttributeVal
		err := child.Start(ctx, consumerFactory, partitions, partitionMap)
		if err != nil {
			return fmt.Errorf("failed to start child %s: %w", child.Path, err)
		}
	}

	n.logger.Info("Started node")
	return nil
}

func (n *QueueTreeNode) Stop(ctx context.Context) error {
	n.logger.Info("Stopping node")

	if n.Dispatcher != nil { // leaf node
		return n.Dispatcher.Stop(ctx)
	}

	for _, child := range n.Children {
		if err := child.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop child %s: %w", child.Path, err)
		}
	}

	n.logger.Info("Stopped node")
	return nil
}

func (n *QueueTreeNode) Enqueue(
	ctx context.Context,
	item types.Item,
	partitions []string,
	partitionMap map[string]any,
) (types.ItemToPersist, error) {
	// If there are no children then this is a leaf node
	if len(n.Children) == 0 {
		return types.NewItemToPersist(item, types.NewItemPartitions(partitions, partitionMap)), nil
	}

	// Add the attribute value to queueNodePathParts
	partitionVal := item.GetAttribute(n.PartitionKey)
	partitions = append(partitions, n.PartitionKey)
	partitionMap[n.PartitionKey] = partitionVal

	child, ok := n.Children[partitionVal]
	if !ok {
		// TODO: thread safety missing
		child, ok = n.Children["*"]
		partitionMap[n.PartitionKey] = "*"
		if !ok {
			// catch-all nodes are created during initalization so this should never happen
			return nil, fmt.Errorf("no child found for attribute %v in node %v", partitionVal, n.Path)
		}
	}

	return child.Enqueue(ctx, item, partitions, partitionMap)
}

func (n *QueueTreeNode) String() string {
	return fmt.Sprintf("QueueTreeNode{Path: %q, AttributeKey: %v, AttributeVal: %v, NodePolicy: %s, Num Children: %d}", n.Path, n.AttributeKey, n.AttributeVal, n.NodePolicy, len(n.Children))
}

func (n *QueueTreeNode) Init(logger log.Logger, scope metrics.Scope, policyCol types.NodePolicyCollection, partitions []string) error {
	n.originalLogger = logger
	n.logger = logger.WithTags(tag.ComponentMapQTreeNode, tag.Dynamic("path", n.Path))
	n.scope = scope

	// Get the merged policy for this node
	policy, err := policyCol.GetMergedPolicyForNode(n.Path)
	if err != nil {
		return err
	}
	n.NodePolicy = policy

	// Set partition key of the node
	nodeLevel := nodeLevel(n.Path)
	if nodeLevel < len(partitions) {
		n.PartitionKey = partitions[nodeLevel]
	}

	// Create predefined children nodes
	return n.addPredefinedSplits(policyCol, partitions)
}

func (n *QueueTreeNode) addChild(attrVal any, policyCol types.NodePolicyCollection, partitions []string) (*QueueTreeNode, error) {
	path := fmt.Sprintf("%s/%v", n.Path, attrVal)
	ch := &QueueTreeNode{
		Path:         path,
		AttributeKey: n.PartitionKey,
		AttributeVal: attrVal,
		Children:     map[any]*QueueTreeNode{},
	}

	if err := ch.Init(n.originalLogger, n.scope, policyCol, partitions); err != nil {
		return nil, err
	}

	n.Children[attrVal] = ch
	return ch, nil
}

func (n *QueueTreeNode) addPredefinedSplits(policyCol types.NodePolicyCollection, partitions []string) error {
	if n.NodePolicy.SplitPolicy == nil || len(n.NodePolicy.SplitPolicy.PredefinedSplits) == 0 {
		return nil
	}

	if nodeLevel(n.Path) >= len(partitions) {
		return fmt.Errorf("predefined split is defined for a leaf level node %s", n.Path)
	}

	for _, split := range n.NodePolicy.SplitPolicy.PredefinedSplits {

		_, err := n.addChild(split, policyCol, partitions)
		if err != nil {
			return err
		}
	}

	return nil
}
