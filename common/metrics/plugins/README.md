# Overview
In Cadence Metric configuration, M3, Prometheus and Statsd are supported natively.
Besides, cadence has a plugin model to implement any metric reporter to emit server metrics to any metrics storage.
The interface `MetricReporterPlugin` is in this package. As an example, Statsd is also implemented as plugin here.

# Load a plugin
To use a plugin, one line of code is needed to any place, usually in main function:
```go
//in import section
_ "github.com/uber/cadence/common/metrics/plugins/statsd"             // needed to load statsd plugin
```   
And re-compile the server binary.

For our example of Statsd, we have added in our main function under `cmd/server/main.go`

# Implement a plugin
Like said in above, any struct that hs implemented the `MetricReporterPlugin` interface could be used as a metric reporter plugin.
You can follow the example of Statsd to implement one. 

It's worth mentioning here that `MetricReporterPlugin` interface is expecting to return a `tally.Scope`.
A tally scope requires to implement a tally Reporter.
There are usually two ways of implementing a tally Reporter:
* Implement a `StatsReporter`. This is an easier [way](https://github.com/uber-go/tally/blob/164eb6a3c0d4c8c82e5a63a8a12d506f0c9b2637/reporter.go#L35).
* Implement a `CachedStatsReporter`. [CachedStatsReporter](https://github.com/uber-go/tally/blob/164eb6a3c0d4c8c82e5a63a8a12d506f0c9b2637/reporter.go#L82) is a backend for Scopes that pre allocates all counter, gauges, timers & histograms. This is harder to implement but more performant. 
 