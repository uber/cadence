// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package history

import (
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

const (
	reapplyCacheMaxSize = 2 * 1024
)

type (
	reapplyCache interface {
		updateReapplyEvent(
			domainID string,
			runID string,
			lastEventID int64,
			lastEventVersion int64,
		) error
		isEventReapplied(
			domainID string,
			runID string,
			lastEventID int64,
			lastEventVersion int64,
		) bool
	}

	reapplyCacheImpl struct {
		cache.Cache

		metricsClient metrics.Client
		logger        log.Logger
	}
)

func newReapplyCache(
	metricsClient metrics.Client,
	logger log.Logger,
) reapplyCache {
	return &reapplyCacheImpl{
		Cache:         cache.NewLRU(reapplyCacheMaxSize),
		metricsClient: metricsClient,
		logger:        logger,
	}
}

func (r *reapplyCacheImpl) updateReapplyEvent(
	domainID string,
	runID string,
	lastEventID int64,
	lastEventVersion int64,
) error {

	key := definition.NewReapplyIdentifier(
		domainID,
		runID,
		lastEventID,
		lastEventVersion)
	_, err := r.PutIfNotExist(key, true)
	if err != nil {
		r.metricsClient.IncCounter(metrics.ReapplyCacheUpdateScope, metrics.CacheFailures)
		return err
	}
	return nil
}

func (r *reapplyCacheImpl) isEventReapplied(
	domainID string,
	runID string,
	lastEventID int64,
	lastEventVersion int64,
) bool {

	key := definition.NewReapplyIdentifier(
		domainID,
		runID,
		lastEventID,
		lastEventVersion)
	_, cacheHit := r.Get(key).(bool)

	return cacheHit
}
