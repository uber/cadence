package sysworkflow

import (
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/metrics"
	"time"
)

type replayMetricsClient struct {
	client   metrics.Client
	isReplay bool
}

// NewReplayMetricsClient creates a metrics client which is aware of cadence's replay mode
func NewReplayMetricsClient(client metrics.Client, isReplay bool) metrics.Client {
	return &replayMetricsClient{
		client:   client,
		isReplay: isReplay,
	}
}

// IncCounter increments a counter metric
func (r *replayMetricsClient) IncCounter(scope int, counter int) {
	if r.isReplay {
		return
	}
	r.client.IncCounter(scope, counter)
}

// AddCounter adds delta to the counter metric
func (r *replayMetricsClient) AddCounter(scope int, counter int, delta int64) {
	if r.isReplay {
		return
	}
	r.client.AddCounter(scope, counter, delta)
}

// StartTimer starts a timer for the given metric name. Time will be recorded when stopwatch is stopped.
func (r *replayMetricsClient) StartTimer(scope int, timer int) tally.Stopwatch {
	if r.isReplay {
		return r.nopStopwatch()
	}
	return r.client.StartTimer(scope, timer)
}

// RecordTimer starts a timer for the given metric name
func (r *replayMetricsClient) RecordTimer(scope int, timer int, d time.Duration) {
	if r.isReplay {
		return
	}
	r.RecordTimer(scope, timer, d)
}

// UpdateGauge reports Gauge type absolute value metric
func (r *replayMetricsClient) UpdateGauge(scope int, gauge int, value float64) {
	if r.isReplay {
		return
	}
	r.UpdateGauge(scope, gauge, value)
}

// Tagged returns a client that adds the given tags to all metrics
func (r *replayMetricsClient) Tagged(tags map[string]string) metrics.Client {
	return &replayMetricsClient{
		client:   r.client.Tagged(tags),
		isReplay: r.isReplay,
	}
}

type nopStopwatchRecorder struct{}

// RecordStopwatch is a nop impl for replay mode
func (n *nopStopwatchRecorder) RecordStopwatch(stopwatchStart time.Time) {}

func (r *replayMetricsClient) nopStopwatch() tally.Stopwatch {
	return tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})
}
