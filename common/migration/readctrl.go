package migration

import (
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"time"
)

type readerImpl[T any] struct {
	log                log.Logger
	scope              metrics.Scope
	operation          string
	backgroundTimeout  time.Duration
	rolloutCtrl        dynamicconfig.StringPropertyFn
	doResultComparison dynamicconfig.BoolPropertyFn
	comparisonFn       ComparisonFn[*T]
}

func NewDualReaderWithCustomComparisonFn[T any](
	rolloutCtrl dynamicconfig.StringPropertyFn,
	doResultComparison dynamicconfig.BoolPropertyFn,
	log log.Logger,
	scope metrics.Scope,
	backgroundTimeout time.Duration,
	comparisonFn ComparisonFn[*T],
) Reader[T] {

	return &readerImpl[T]{
		log:                log,
		scope:              scope,
		rolloutCtrl:        rolloutCtrl,
		doResultComparison: doResultComparison,
		backgroundTimeout:  backgroundTimeout,
		comparisonFn:       comparisonFn,
	}
}

func (c *readerImpl[T]) getReaderRolloutState(constraints Constraints) ReaderRolloutState {
	s := c.rolloutCtrl(
		dynamicconfig.OperationFilter(c.operation),
		dynamicconfig.DomainFilter(constraints.Domain),
	)
	return ReaderRolloutState(s)
}

func (c *readerImpl[T]) shouldCompare(constraints Constraints) bool {
	return c.doResultComparison(
		dynamicconfig.OperationFilter(c.operation),
		dynamicconfig.DomainFilter(constraints.Domain),
	)
}
