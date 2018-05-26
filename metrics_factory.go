package simetrics

import (
	"github.com/sirupsen/logrus"

	"github.com/luismfonseca/simetrics/simetricsconfig"
	"github.com/luismfonseca/simetrics/sink"
)

// Factory method to build `SiMetrics` from a config
func FromConfig(conf *simetricsconfig.Config, log *logrus.Entry) *SiMetrics {
	mOpts := MetricsOptions{TrackVarsPeriod: conf.TrackVarsPeriod, NamespaceFormat: conf.NamespaceFormat}
	mBuilder := NewBuilder(mOpts, sink.FromConfig(conf, log))

	m, err := mBuilder.Build()
	if err != nil {
		log.WithError(err).Warn("Failed to init the metrics. Not sending any metrics...")
		return NewEmpty()
	}

	return m
}
