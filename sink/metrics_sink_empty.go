package sink

// Doesn't do anything, very efficient
type MetricsSinkEmpty struct{}

func (mse MetricsSinkEmpty) Init() error {
	return nil
}

func (mse MetricsSinkEmpty) ReportCount(name string, value float64) {}

func (mse MetricsSinkEmpty) ReportValue(name string, value float64) {}

func (mse MetricsSinkEmpty) ReportDistribution(name string, value float64) {}
