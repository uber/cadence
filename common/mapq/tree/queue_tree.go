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
	"strings"

	"github.com/uber/cadence/common/mapq/types"
)

// QueueTree is a tree structure that represents the queue structure for MAPQ
type QueueTree struct {
	partitions      []string
	policyCol       types.NodePolicyCollection
	persister       types.Persister
	consumerFactory types.ConsumerFactory
	root            *QueueTreeNode
}

func New(
	partitions []string,
	policies []types.NodePolicy,
	persister types.Persister,
	consumerFactory types.ConsumerFactory,
) (*QueueTree, error) {
	t := &QueueTree{
		partitions:      partitions,
		policyCol:       types.NewNodePolicyCollection(policies),
		persister:       persister,
		consumerFactory: consumerFactory,
	}

	err := t.init()
	return t, err
}

// Start the dispatchers for all leaf nodes
func (t *QueueTree) Start(ctx context.Context) error {
	return t.root.Start(ctx, t.consumerFactory, nil, map[string]any{})
}

// Stop the dispatchers for all leaf nodes
func (t *QueueTree) Stop(ctx context.Context) error {
	return t.root.Stop(ctx)
}

func (t *QueueTree) String() string {
	var sb strings.Builder
	var nodes []*QueueTreeNode
	nodes = append(nodes, t.root)
	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]
		sb.WriteString(node.String())
		sb.WriteString("\n")
		for _, child := range node.Children {
			nodes = append(nodes, child)
		}
	}

	return sb.String()
}

func (t *QueueTree) Enqueue(ctx context.Context, items []types.Item) error {
	if t.root == nil {
		return fmt.Errorf("root node is nil")
	}

	var itemsToPersist []types.ItemToPersist
	for _, item := range items {
		itemToPersist, err := t.root.Enqueue(ctx, item, nil, map[string]any{})
		if err != nil {
			return err
		}
		itemsToPersist = append(itemsToPersist, itemToPersist)
	}

	return t.persister.Persist(ctx, itemsToPersist)
}

func (t *QueueTree) init() error {
	t.root = &QueueTreeNode{
		Path:     "*", // Root node
		Children: map[any]*QueueTreeNode{},
	}

	if err := t.root.Init(t.policyCol, t.partitions); err != nil {
		return fmt.Errorf("failed to initialize root node: %w", err)
	}

	// Create tree nodes with catch-all nodes at all levels and predefined splits.
	// There will be len(partitions) levels in the tree.
	err := t.addCatchAllNodes(t.root)
	if err != nil {
		return fmt.Errorf("failed to initialize tree: %w", err)
	}

	return nil
}

func (t *QueueTree) addCatchAllNodes(n *QueueTreeNode) error {
	nodeLevel := nodeLevel(n.Path)
	if nodeLevel == len(t.partitions) { // reached the leaf level
		return nil
	}

	if n.Children["*"] != nil { // catch-all node already exists
		return nil
	}

	_, err := n.addChild("*", t.policyCol, t.partitions)
	if err != nil {
		return err
	}

	for _, child := range n.Children {
		if err := t.addCatchAllNodes(child); err != nil {
			return err
		}
	}

	return nil
}

func nodeLevel(path string) int {
	return len(strings.Split(path, "/")) - 1
}
