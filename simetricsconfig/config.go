package simetricsconfig

import (
	"time"
)

type LibratoConfig struct {
	Email        string `mapstructure:"email"`
	Token        string `mapstructure:"token"`
	Namespace    string `mapstructure:"namespace"`
	SourceFormat string `mapstructure:"source-format"` // interpolated with the hostname
}

type Config struct {
	Backend         string         `mapstructure:"backend"`
	Librato         *LibratoConfig `mapstructure:"librato"`
	NamespaceFormat string         `mapstructure:"namespace-format"`
	TrackVarsPeriod time.Duration  `mapstructure:"track-vars-period"` // defaults to 5s
}
