package internal

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/quotas"
)

func TestShadowed(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		tests := map[string]struct {
			primary bool
			shadow  bool
			allowed bool
			pUsage  UsageMetrics
			sUsage  UsageMetrics
		}{
			"only primary allows": {
				primary: true,
				shadow:  false,
				allowed: true,
				pUsage:  UsageMetrics{Allowed: 1},
				sUsage:  UsageMetrics{Rejected: 1},
			},
			"only shadow allows": {
				primary: false,
				shadow:  true,
				allowed: false,
				pUsage:  UsageMetrics{Rejected: 1},
				sUsage:  UsageMetrics{Allowed: 1},
			},
			"both allow": {
				primary: true,
				shadow:  true,
				allowed: true,
				pUsage:  UsageMetrics{Allowed: 1},
				sUsage:  UsageMetrics{Allowed: 1},
			},
			"both reject": {
				primary: false,
				shadow:  false,
				allowed: false,
				pUsage:  UsageMetrics{Rejected: 1},
				sUsage:  UsageMetrics{Rejected: 1},
			},
		}
		for name, test := range tests {
			test := test
			t.Run("allow-"+name, func(t *testing.T) {
				ts := clock.NewMockedTimeSource()
				primaryBurst, shadowBurst := 0, 0
				if test.primary {
					primaryBurst = 1
				}
				if test.shadow {
					shadowBurst = 1
				}
				primary := NewCountedLimiter(clock.NewMockRatelimiter(ts, rate.Every(time.Second), primaryBurst))
				shadow := NewCountedLimiter(clock.NewMockRatelimiter(ts, rate.Every(time.Second), shadowBurst))
				s := NewShadowedLimiter(primary, shadow)

				assert.Equalf(t, test.allowed, s.Allow(), "should match primary behavior: %v", test.allowed)
				assert.Equal(t, test.pUsage, primary.Collect(), "should have called primary")
				assert.Equal(t, test.sUsage, shadow.Collect(), "should have called shadow")
			})
			t.Run("reserve-"+name, func(t *testing.T) {
				ts := clock.NewMockedTimeSource()
				primaryBurst, shadowBurst := 0, 0
				if test.primary {
					primaryBurst = 1
				}
				if test.shadow {
					shadowBurst = 1
				}
				primary := NewCountedLimiter(clock.NewMockRatelimiter(ts, rate.Every(time.Second), primaryBurst))
				shadow := NewCountedLimiter(clock.NewMockRatelimiter(ts, rate.Every(time.Second), shadowBurst))
				s := NewShadowedLimiter(primary, shadow)

				res := s.Reserve()
				assert.Equalf(t, test.allowed, res.Allow(), "should match primary behavior: %v", test.allowed)
				res.Used(true)
				assert.Equal(t, test.pUsage, primary.Collect(), "should have called primary")
				assert.Equal(t, test.sUsage, shadow.Collect(), "should have called shadow")
			})
			t.Run("wait-"+name, func(t *testing.T) {
				// the Wait operation is inherently racy in how it calls both limiters,
				// so this test delays the primary's response a bit to stabilize behavior.
				ctrl := gomock.NewController(t)
				primary := quotas.NewMockLimiter(ctrl)
				primary.EXPECT().Wait(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
					select {
					case <-ctx.Done(): // canceled
						return ctx.Err()
					case <-time.After(10 * time.Millisecond): // give the shadow some time to run
						if test.primary {
							return nil
						}
						return context.Canceled
					}
				})
				cPrimary := NewCountedLimiter(primary)
				shadow := quotas.NewMockLimiter(ctrl)
				shadow.EXPECT().Wait(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
					// shadow returns immediately to keep the test's inherent race
					// more stable.
					if test.shadow {
						return nil
					}
					return context.Canceled
				})
				cShadow := NewCountedLimiter(shadow)
				s := NewShadowedLimiter(cPrimary, cShadow)

				requireQuickly(t, 100*time.Millisecond, func() {
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						time.Sleep(50 * time.Millisecond)
						cancel()
					}()
					err := s.Wait(ctx)
					if test.allowed {
						assert.NoError(t, err, "should have waited until allowed")
					} else {
						assert.Error(t, err, "should have been canceled")
					}
				})
				assert.Equal(t, test.pUsage, cPrimary.Collect(), "should have called primary")
				assert.Equal(t, test.sUsage, cShadow.Collect(), "should have called shadow")
			})
		}
	})
	t.Run("reserve", func(t *testing.T) {
		ts := clock.NewMockedTimeSource()
		primary := NewCountedLimiter(clock.NewMockRatelimiter(ts, 1, 1)) // allows an event
		shadow := NewCountedLimiter(clock.NewMockRatelimiter(ts, 0, 0))  // always rejects
		s := NewShadowedLimiter(primary, shadow)

		assert.True(t, s.Allow(), "should match primary behavior")
		assert.Equal(t, UsageMetrics{1, 0, 0}, primary.Collect(), "should have called primary")
		assert.Equal(t, UsageMetrics{0, 1, 0}, shadow.Collect(), "should have called shadow")
	})
}
