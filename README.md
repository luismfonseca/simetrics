# SMetrics

Service Metrics library exposes an interface that makes emitting metrics as simple as doing a log line.
Currently supports Librato as a backend.

## Examples

Initializing the library:

```
metric := smetrics.NewEmpty()
if libratoToken != "" {
	metric := MetricsFromConfig(
		&metricsconfig.MetricsConfig{
			Backend: "librato",
			Librato: &metricsconfig.MetricsLibratoConfig{
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


    // or if there's a mutable variable:
    metric.TrackVarInt("producer.state", &prodState)
    metric.TrackVarFloat("last_seen.timestamp", &lastDevice.timestamp)
```
