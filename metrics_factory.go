package smetrics

import (
	"github.com/sirupsen/logrus"

	. "github.com/luismfonseca/smetrics/config"
	"github.com/luismfonseca/smetrics/sink"
)

// Factory method to build `SMetrics` from a config
func MetricsFromConfig(conf *MetricsConfig, log *logrus.Entry) *SMetrics {
	mOpts := MetricsOptions{TrackVarsPeriod: conf.TrackVarsPeriod, NamespaceFormat: conf.NamespaceFormat}
	mBuilder := NewBuilder(mOpts, sink.FromConfig(conf, log))

	m, err := mBuilder.Build()
	if err != nil {
		log.WithError(err).Warn("Failed to init the metrics. Not sending any metrics...")
		return NewEmpty()
	}

	return m
}
