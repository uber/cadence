package mapq

import (
	"context"

	"github.com/uber/cadence/playground/mapq/tree"
	"github.com/uber/cadence/playground/mapq/types"
)

type ClientImpl struct {
	persister       types.Persister
	consumerFactory types.ConsumerFactory
	tree            *tree.QueueTree
	partitions      []string
	policies        []types.NodePolicy
}

func (c *ClientImpl) Start(context.Context) error {
	return nil
}

func (c *ClientImpl) Stop(context.Context) error {
	return nil
}

func (c *ClientImpl) Enqueue(context.Context, []types.Item) error {
	return nil
}

func (c *ClientImpl) Ack(context.Context, types.Item) error {
	return nil
}

func (c *ClientImpl) Nack(context.Context, types.Item) error {
	return nil
}
