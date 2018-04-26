package sink

type MetricsSink interface {
	// Performs any necessary initialization, returning an error if something fails
	Init() error

	// Reports a count (this is a delta value)
	ReportCount(name string, value float64)

	// Reports the current value
	ReportValue(name string, value float64)

	// Reports another value for a distribution
	ReportDistribution(name string, value float64)
}
