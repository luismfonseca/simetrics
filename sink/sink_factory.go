package sink

import (
	"github.com/luismfonseca/simetrics/simetricsconfig"
	"github.com/sirupsen/logrus"
)

func FromConfig(config *simetricsconfig.Config, log *logrus.Entry) MetricsSink {
	switch config.Backend {
	case "librato":
		log.WithField("backend", "librato").Info("Using 'librato' backend for metrics.")
		return NewMetricsSinkLibrato(
			config.Librato.Email,
			config.Librato.Token,
			config.Librato.Namespace,
			config.Librato.SourceFormat,
			log,
		)
	case "stdout":
		return NewMetricsSinkStdout(log)
	case "none", "empty":
		log.WithField("backend", config.Backend).Info("Metrics reporting is explicitly disabled.")
		return &MetricsSinkEmpty{}
	default:
		log.WithField("backend", config.Backend).Warn("Undefined or unknown backend. Not sending any metrics...")
		return &MetricsSinkEmpty{}
	}
}
