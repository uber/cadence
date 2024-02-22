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
Package global will contain a global-load-balance-aware ratelimiter.

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

The weight-calculation algorithm will (eventually) be controlled by a dynamic config value,
to allow shadowing and experimenting with different algorithms at runtime, or disabling
the global logic altogether.

See sub-packages for implementation details.
*/
package global
