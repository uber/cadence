package mapq

import (
	"fmt"

	"github.com/uber/cadence/playground/mapq/tree"
	"github.com/uber/cadence/playground/mapq/types"
)

type Options func(*ClientImpl)

func WithPersister(p types.Persister) Options {
	return func(c *ClientImpl) {
		c.persister = p
	}
}

func WithConsumerFactory(cf types.ConsumerFactory) Options {
	return func(c *ClientImpl) {
		c.consumerFactory = cf
	}
}

// WithPartitions sets the partition keys for each level.
// MAPQ creates a tree with depth = len(partitions)
func WithPartitions(partitions []string) Options {
	return func(c *ClientImpl) {
		c.partitions = partitions
	}
}

// WithPolicies sets the policies for the MAPQ instance.
// Policies can be defined for nodes at a specific level or nodes with specific path.
//
// Path conventions:
// - "" -> empty path represents the root node at level 0
// - "/." matches with all nodes at level 1
// - "/*" represents the catch-all node at level 1
// - "/xyz" represents a specific node at level 1 whose partition value is xyz
// - "/./." matches with all nodes at level 2
// - "/xyz/." matches with all nodes at level 2 whose parent is xyz node
// - "xyz/*" represents the catch-all node at level 2 whose parent is xyz node
// - "xyz/abc" represents a specific node at level 2 whose level 2 attribute value is abc and parent is xyz node
func WithPolicies(policies []types.NodePolicy) Options {
	return func(c *ClientImpl) {
		c.policies = policies
	}
}

func New(opts ...Options) (types.Client, error) {
	c := &ClientImpl{}

	for _, opt := range opts {
		opt(c)
	}

	if c.persister == nil {
		return nil, fmt.Errorf("persister is required. Use WithPersister option to set it")
	}

	if c.consumerFactory == nil {
		return nil, fmt.Errorf("consumer factory is required. Use WithConsumerFactory option to set it")
	}

	tree, err := tree.New(c.partitions, c.policies, c.persister)
	if err != nil {
		return nil, err
	}

	c.tree = tree
	fmt.Printf("tree: \n%s\n", c.tree)

	return c, nil
}
