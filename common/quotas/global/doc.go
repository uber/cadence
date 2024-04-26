// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

/*
Package global contains a global-load-balance-aware ratelimiter (when complete).

# High level overview

At a very high level, this will:
  - collect usage metrics in-memory in "limiting" hosts (which use an in-memory ratelimiter)
  - asynchronously submit these usage metrics to "aggregating" hosts, which will return per-key RPS data
  - "limiting" hosts use this returned value to update their in-memory ratelimiters

"Aggregating" hosts will internally:
  - receive data for key X from all "limiting" hosts
  - compute request-weight per host+key in the cluster
  - return the "limiting"-host's weighted portion of each ratelimit to allow

The exact logic that the "aggregating" hosts use is intentionally hidden from the
"limiting" hosts, so it can be changed without changing how they enforce limits.

The weight-calculation algorithm will (eventually) be controlled by a dynamicconfig value,
to allow shadowing and experimenting with different algorithms at runtime, or disabling
the global logic altogether.

See sub-packages for implementation details.

# Data flow in the system

The overall planned system's data-flow is driven entirely by limiting hosts repeatedly
calling an "update(request, data)" API on aggregating hosts, and using the return value.
Aggregating hosts make no outbound requests at all:

	lim: a host limiting e.g. user RPS for domain ASDF
	agg: a host aggregating this data, to compute fair RPS allowances
	┌───┐              ┌────┐┌────┐┌────┐              ┌────┐
	│lim│              │ring││agg1││agg2│              │lim2│
	└─┬─┘              └─┬──┘└─┬──┘└─┬──┘              └─┬──┘
	  │                  │     │     │                   │
	  │ shard(a,b,c)     │     │     │    update(a1,d37) │
	  │────────────────>┌┴┐   ┌┴┐<───────────────────────│
	  │    [[a], [b,c]] │ │   └┬┘───────────────────────>│
	  │<────────────────└┬┘    │     │                  ...
	  │                  │     │     │
	  │ update(a1)       │     │     │   got 1 request on "a" limit
	  │──────────────────────>┌┴┐    │
	  │                  │    │ │    │
	  │ update(b1,c2)    │    │ │    │   1 req on "b", 2 on "c"
	  │────────────────────────────>┌┴┐
	  │                  │    │ │   │ │
	  │     [b4/s, c9/s] │    │ │   │ │  allow "b" 4 RPS, "c" gets 9
	  │<────────────────────────────└┬┘
	  │                  │    │ │    │
	  │           [a2/s] │    │ │    │   2 RPS for "a"
	  │<──────────────────────└┬┘    │
	  │                  │     │     │
	┌─┴─┐              ┌─┴──┐┌─┴──┐┌─┴──┐
	│lim│              │ring││agg1││agg2│
	└───┘              └────┘└────┘└────┘

This package as a whole is only concerned with the high level request pattern in use:
  - limiters enforce limits and collect data
  - limiters asynchronously submit that data to all relevant aggregators, sharded by key
  - aggregators decide per-limiter-and-key limits based on global data, and return it to limiters

Currently:
  - limiters are Frontends, but can trivially be any Cadence service as they only send outbound requests
  - aggregators are arbitrarily in History service because it already has a ring implemented (Frontend does not)
  - neither of these are fundamentally required, and they may change or be in multiple services later

The actual contents of the requests and responses and how limits are decided is
intentionally flexible to allow creating custom algorithms / data / etc externally without
requiring internal RPC spec definitions.  To achieve this, we have a custom
[github.com/uber/cadence/common/types.Any] type that can be used with either Thrift or Protobuf.
This type is intentionally not google.protobuf.Any because there is no reason to require
exchanging Protobuf data (particularly if using Thrift), and because that tends to bind to a
specific code generator, but it serves essentially the same purpose.

A built-in version of all this will be coming in later commits, and possibly more in the future
if we end up needing major changes that are reusable.

# Planned use

The first use of this is intentionally focused on addressing imbalanced usage
of per-domain user/worker/ES ratelimits across different frontend hosts, e.g.
due to unfair load balancing or per-datacenter affinity or other
occasionally-changing-else-mostly-stable routing decisions.
This will essentially imply 3 keys per domain, and will look something like
`domain-user-request` / `domain-worker-request` / etc - they just have to be
unique per conceptual thing-with-a-configured-limit.

These frontend limits are generally configured up-front and are used to protect
the cluster (especially the database) from noisy neighbors and un-planned major
usage changes.  True precision is not necessary (accepting 110 instead of 100 will
not collapse the cluster, but allowing *unlimited* could), but users plan around the
numbers we have given them, and significant or sustained differences in reality can
cause problems for them.  Particularly if it significantly decreases due to a change
in how our network decides to distribute their requests.

By detecting this imbalanced load we can fairly allow "hot" hosts more RPS than our
"limit / N hosts" limiters allow (essentially stealing unused quota from low-usage
hosts), and ultimately allow request rates closer to the intended limits as the
balance changes over time.  It's better for our users (they get what we told them
they get), and better for us (no need to artificially raise the limit to work around
the load imbalance, nor keep changing it as load changes).

Weight-data loss is expected due to frontend or agg process losses, and allowed-limits
naturally trail real-time use due to the asynchronous nature - this is not *ideal* but
is expected to be acceptable, as it will adjust / self-repair "soon" as normal updates
occur.

More concretely, our planned target is something like:
  - frontends (the limiting hosts) submit usage data every 3s
  - history's aggregating algorithms "quickly" account for new load, and return allowed
    limits to frontends
  - within about 10s of a change, we want things to be ~90% accurate or better
    compared to the "ground truth" (this implies ~0.5 weight per 3s update)
  - with stable traffic, >99% accuracy at allowing the intended RPS of a limit key,
    regardless of how requests are distributed
  - briefly going over or under the intended limit during changes or partial data
    is allowed, but generally requests should be incorrectly rejected over
    incorrectly allowed

Future uses (with minor changes) may include things like:
  - limiting cluster-wide database RPS across all hosts (one key, 50k+ RPS) despite
    significantly-different load depending on server role and shard distribution
  - different load-deciding algorithms, e.g. TCP Vegas or some other more sophisticated
    weight-deciding algorithm (this first one is intentionally simple, and might prove
    good enough)
  - closed-source pluggable algorithms that share the same intent and request-flow
    but little else (i.e. extendable by operators without new RPC definitions.
    routing and usage patterns are often unique, no one approach can be ideal everywhere)

# Out of scope

Uses explicitly out of scope for this whole structure includes things like:
  - perfect accuracy on RPS targets (requires synchronous decisions)
  - rapid and significant usage changes, like a domain sending all requests to one
    frontend host one second, then a different the next, etc.
  - brief spikes (e.g. sub-10-seconds.  these are broadly undesirable and can
    behave worse than RPS/hosts would allow)
  - guaranteed limiting below intended RPS (partial data may lead to split-brain-like
    behavior, leading to over-allowing)
  - fully automated target-RPS adjusting (e.g. system-load detecting, not pre-configured RPS.
    this may be *possible* to implement in this system, but it is not *planned* and
    may require significant changes)
*/
package global
