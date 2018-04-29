# SiMetrics

This library exposes an interface that makes emitting metrics as simple as doing a log line.
Currently supports Librato as a backend.

SiMetrics? Por supuesto que sí.

## Examples

Initializing the library:

```
metric := simetrics.NewEmpty()
if libratoToken != "" {
	metric := simetrics.FromConfig(
		&config.MetricsConfig{
			Backend: "librato",
			Librato: &metrics.MetricsLibratoConfig{
				Email:        "email@somewhere.com",
				Token:        libratoToken,
				Namespace:    "product-a",
				SourceFormat: "staging_%s",
			},
			NamespaceFormat: "%s.",
		},
		logrus.New().WithField("service", "product-a"),
	)
}
```

Primitives:
```
metric.Increment("new_device.error.4xx.invalid_body")
metric.Decrement("active_connections")
metric.Count("recipient_fanout", float64(len(recipients)))
metric.Value("speed", currentSpeed()) // just like a gauge
```

Distributions:
```
metric.Distribution("new_device.json_body_size_bytes", float64(len(requestBodyBytes)))
```

If you want to time a function:

```
func handleNewDevice() {
    timeStart := time.Now()
    defer metric.TimeSince("new_device.latency_ms", timeStart)
    
    // If you prefer the one liner you can just do:
    defer metric.TimeSince("new_device.latency_ms", time.Now())
}
```

There are some helper functions to periodically track values:

```
// By default, this metric will be emitted every 5s:
metric.TrackFuncInt("runtime.num_goroutines", runtime.NumGoroutine)

metric.TrackFuncFloat("db.connection_pool_usage", func() float64 {
    numOpenConnections := float64(underlyingDB.Stats().OpenConnections)
    return numOpenConnections / float64(maxOpenConnections)
})


// You can also stop tracking metrics by storing the return value:
lastPeriodStartedMetric := metric.TrackFuncInt("aggregator.period_start_seconds_ago", func() int {
    lastUpdatedTime := time.Unix(enf.agg.periodStartTime, 0)

    return int(time.Since(lastUpdatedTime).Seconds())
})

defer lastPeriodStartedMetric.Stop()

// Or by stopping all:
metrics.StopAllTrackingMetrics()
```
