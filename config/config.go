package metricsconfig

import (
	"time"
)

type MetricsStaggerConfig struct {
	Address string `mapstructure:"address"`
}

type MetricsLibratoConfig struct {
	Email        string `mapstructure:"email"`
	Token        string `mapstructure:"token"`
	Namespace    string `mapstructure:"namespace"`
	SourceFormat string `mapstructure:"source-format"` // interpolated with the hostname
}

type MetricsConfig struct {
	Backend         string                `mapstructure:"backend"`
	Stagger         *MetricsStaggerConfig `mapstructure:"stagger"`
	Librato         *MetricsLibratoConfig `mapstructure:"librato"`
	NamespaceFormat string                `mapstructure:"namespace-format"`
	TrackVarsPeriod time.Duration         `mapstructure:"track-vars-period"` // defaults to 5s
}
