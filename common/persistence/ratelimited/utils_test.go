package ratelimited

import (
	"context"

	"github.com/uber/cadence/common/quotas"
	"golang.org/x/time/rate"
)

type limiterAlwaysAllow struct{}

func (l limiterAlwaysAllow) Allow() bool {
	return true
}

func (l limiterAlwaysAllow) Wait(ctx context.Context) error {
	return nil
}

func (l limiterAlwaysAllow) Reserve() *rate.Reservation {
	return &rate.Reservation{}
}

type limiterNeverAllow struct{}

func (l limiterNeverAllow) Allow() bool {
	return false
}

func (l limiterNeverAllow) Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (l limiterNeverAllow) Reserve() *rate.Reservation {
	return &rate.Reservation{}
}

var _ quotas.Limiter = (*limiterAlwaysAllow)(nil)
var _ quotas.Limiter = (*limiterNeverAllow)(nil)
