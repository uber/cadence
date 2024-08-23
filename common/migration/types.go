package migration

import (
	"context"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type ReaderRolloutState string
type WriterRolloutState string

// A function for doing comparisons on data returned. Returns true if they're equal
type ComparisonFn[T any] func(log.Logger, metrics.Scope, T, error, T, error) bool

type Constraints struct {
	Operation string
	Domain    string
}

type Reader[T any] interface {
	ReadAndReturnActive(
		ctx context.Context,
		constraints Constraints,
		old func(ctx context.Context) (*T, error),
		new func(ctx context.Context) (*T, error),
	) (*T, error)
}
