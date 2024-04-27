/*
Package limiter contains the limiting-host ratelimit usage tracking and enforcing logic.

At a very high level, this wraps a quotas.Limiter to do a few additional things
in the context of the [github.com/uber/cadence/common/quotas/global] ratelimiter system:
- keep track of usage per key (quotas.Limiter does not support this natively, nor should it)
- periodically report usage to an "aggregator" host
- apply the aggregator's returned per-key RPS limits to future requests
- fall back to the wrapped limiter in case of failures
*/
package limiter

import _ "github.com/uber/cadence/common/quotas"

// GlobalLimiter is a restricted version of quotas.Limiter for use with the global
// ratelimiter system, because features like `Wait` are difficult to support in an
// obviously-reasonable way, and they are not currently needed.
type GlobalLimiter interface {
	Allow() bool
}
