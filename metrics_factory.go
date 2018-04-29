package simetrics

import (
	"github.com/sirupsen/logrus"

	"github.com/luismfonseca/simetrics/config"
	"github.com/luismfonseca/simetrics/sink"
)

// Factory method to build `SiMetrics` from a config
func FromConfig(conf *config.MetricsConfig, log *logrus.Entry) *SiMetrics {
	mOpts := MetricsOptions{TrackVarsPeriod: conf.TrackVarsPeriod, NamespaceFormat: conf.NamespaceFormat}
	mBuilder := NewBuilder(mOpts, sink.FromConfig(conf, log))

	m, err := mBuilder.Build()
	if err != nil {
		log.WithError(err).Warn("Failed to init the metrics. Not sending any metrics...")
		return NewEmpty()
	}

	return m
}
