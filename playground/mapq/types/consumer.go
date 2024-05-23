package types

import "context"

type ConsumerFactory interface {
	New(context.Context, ItemPartitions) Consumer
}

type Consumer interface {
	Start(context.Context) error
	Stop(context.Context) error
	Process(context.Context, Item) error
}
