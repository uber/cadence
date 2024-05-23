package types

import (
	"context"
)

type Client interface {
	// Enqueue adds an item to the queue
	Enqueue(context.Context, []Item) error

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

type Persister interface {
	Persist(ctx context.Context, items []ItemToPersist) error
	GetOffsets(ctx context.Context) (*Offsets, error)
	CommitOffsets(ctx context.Context, offsets *Offsets) error
}

type Offsets struct {
	// TODO: define offsets
}
