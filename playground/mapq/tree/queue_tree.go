package tree

import (
	"context"
	"fmt"
	"strings"

	"github.com/uber/cadence/playground/mapq/types"
)

type QueueTree struct {
	partitions []string
	policyCol  types.NodePolicyCollection
	persister  types.Persister
	root       *QueueTreeNode
}

type QueueTreeNode struct {
	// The path to the node
	Path string

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
}

func New(partitions []string, policies []types.NodePolicy, p types.Persister) (*QueueTree, error) {
	t := &QueueTree{
		partitions: partitions,
		policyCol:  types.NewNodePolicyCollection(policies),
		persister:  p,
	}

	err := t.init()
	return t, err
}

func (t *QueueTree) init() error {
	rootPolicy, err := t.policyCol.GetMergedPolicyForNode("")
	if err != nil {
		return fmt.Errorf("failed to get root policy: %w", err)
	}
	t.root = &QueueTreeNode{
		Path:       "", // Root node
		Children:   map[any]*QueueTreeNode{},
		NodePolicy: rootPolicy,
	}

	// Create tree nodes with catch-all nodes at all levels and predefined splits.
	// There will be len(partitions) levels in the tree.
	err = t.addCatchAllNodes(t.root)
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

	nextLevelAttrKey := t.partitions[nodeLevel]
	ch, err := n.addChild(nextLevelAttrKey, "*", t.policyCol, t.partitions)
	if err != nil {
		return err
	}

	return t.addCatchAllNodes(ch)
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

func (n *QueueTreeNode) String() string {
	return fmt.Sprintf("QueueTreeNode{Path: %q, AttributeKey: %v, AttributeVal: %v, NodePolicy: %s, Num Children: %d}", n.Path, n.AttributeKey, n.AttributeVal, n.NodePolicy, len(n.Children))
}

func (n *QueueTreeNode) addChild(attrKey string, attrVal any, policyCol types.NodePolicyCollection, partitions []string) (*QueueTreeNode, error) {
	path := fmt.Sprintf("%s/%v", n.Path, attrVal)
	policy, err := policyCol.GetMergedPolicyForNode(path)
	if err != nil {
		return nil, err
	}
	ch := &QueueTreeNode{
		Path:         path,
		AttributeKey: attrKey,
		AttributeVal: attrVal,
		NodePolicy:   policy,
		Children:     map[any]*QueueTreeNode{},
	}

	if err := ch.addPredefinedSplits(policyCol, partitions); err != nil {
		return nil, err
	}

	n.Children[attrVal] = ch
	return ch, nil
}

func (n *QueueTreeNode) addPredefinedSplits(policyCol types.NodePolicyCollection, partitions []string) error {
	if n.NodePolicy.SplitPolicy == nil {
		return nil
	}

	for _, split := range n.NodePolicy.SplitPolicy.PredefinedSplits {
		path := fmt.Sprintf("%s/%v", n.Path, split)
		level := nodeLevel(path)
		if level >= len(partitions) {
			return fmt.Errorf("predefined split is defined for a leaf level node %s", n.Path)
		}

		attrKey := partitions[level]
		_, err := n.addChild(attrKey, split, policyCol, partitions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *QueueTreeNode) Enqueue(ctx context.Context, item types.Item, partitions []string, partitionMap map[string]any) (types.ItemToPersist, error) {
	// Add the attribute value to queueNodePathParts
	attr := item.GetAttribute(n.AttributeKey)
	partitions = append(partitions, n.AttributeKey)
	partitionMap[n.AttributeKey] = item.GetAttribute(n.AttributeKey)

	// If there are no children then this is a leaf node
	if len(n.Children) == 0 {
		return types.NewItemToPersist(item, types.NewItemPartitions(partitions, partitionMap)), nil
	}

	child, ok := n.Children[attr]
	if !ok {
		// TODO: thread safety missing
		child, ok = n.Children["*"]
		if !ok {
			// catch-all nodes are created during initalization so this should never happen
			return nil, fmt.Errorf("no child found for attribute %v in node %v", attr, n.Path)
		}
	}

	return child.Enqueue(ctx, item, partitions, partitionMap)
}

func nodeLevel(path string) int {
	return len(strings.Split(path, "/")) - 1
}
