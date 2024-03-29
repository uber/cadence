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

package metered

// Code generated by gowrap. DO NOT EDIT.
// template: ../templates/metered.tmpl
// gowrap: http://github.com/hexdigest/gowrap

import (
	"context"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

// meteredConfigStoreManager implements persistence.ConfigStoreManager interface instrumented with rate limiter.
type meteredConfigStoreManager struct {
	base
	wrapped persistence.ConfigStoreManager
}

// NewConfigStoreManager creates a new instance of ConfigStoreManager with ratelimiter.
func NewConfigStoreManager(
	wrapped persistence.ConfigStoreManager,
	metricClient metrics.Client,
	logger log.Logger,
	cfg *config.Persistence,
) persistence.ConfigStoreManager {
	return &meteredConfigStoreManager{
		wrapped: wrapped,
		base: base{
			metricClient:                  metricClient,
			logger:                        logger,
			enableLatencyHistogramMetrics: cfg.EnablePersistenceLatencyHistogramMetrics,
		},
	}
}

func (c *meteredConfigStoreManager) Close() {
	c.wrapped.Close()
	return
}

func (c *meteredConfigStoreManager) FetchDynamicConfig(ctx context.Context, cfgType persistence.ConfigType) (fp1 *persistence.FetchDynamicConfigResponse, err error) {
	op := func() error {
		fp1, err = c.wrapped.FetchDynamicConfig(ctx, cfgType)
		c.emptyMetric("ConfigStoreManager.FetchDynamicConfig", cfgType, fp1, err)
		return err
	}

	err = c.call(metrics.PersistenceFetchDynamicConfigScope, op, getCustomMetricTags(cfgType)...)
	return
}

func (c *meteredConfigStoreManager) UpdateDynamicConfig(ctx context.Context, request *persistence.UpdateDynamicConfigRequest, cfgType persistence.ConfigType) (err error) {
	op := func() error {
		err = c.wrapped.UpdateDynamicConfig(ctx, request, cfgType)
		c.emptyMetric("ConfigStoreManager.UpdateDynamicConfig", request, err, err)
		return err
	}

	err = c.call(metrics.PersistenceUpdateDynamicConfigScope, op, getCustomMetricTags(request)...)
	return
}
