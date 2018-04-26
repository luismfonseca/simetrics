package smetrics

import (
	"context"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/luismfonseca/smetrics/sink"
)

type MetricsOptions struct {
	// The period in which variables should be tracked
	TrackVarsPeriod time.Duration

	// A prefix with the app name to fill the `%s`
	NamespaceFormat string

	// computed namespaceFormat + app name
	namespace string
}

type SMetrics struct {
	sink          sink.MetricsSink
	opts          MetricsOptions
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
}

type SMetricsBuilder struct {
	m SMetrics
}

// A tracking metric that can be stopped.
type TrackingMetric context.CancelFunc

func (t TrackingMetric) Stop() {
	t()
}

// Builds a `SMetricsBuilder` so you can call `Build()` to get the built `SMetrics`
func NewSMetricBuilder(options MetricsOptions, ms sink.MetricsSink) *SMetricsBuilder {
	if options.TrackVarsPeriod == 0 {
		options.TrackVarsPeriod = 5 * time.Second
	}

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	return &SMetricsBuilder{m: SMetrics{sink: ms, opts: options, ctx: ctx, ctxCancelFunc: ctxCancelFunc}}
}

// Returns a built and fully initialized `SMetrics` or an error
func (mb *SMetricsBuilder) Build() (*SMetrics, error) {
	err := mb.m.sink.Init()
	if err != nil {
		return nil, err
	}

	if strings.Contains(mb.m.opts.NamespaceFormat, "%s") {
		mb.m.opts.namespace = fmt.Sprintf(mb.m.opts.NamespaceFormat, path.Base(os.Args[0]))
	} else {
		mb.m.opts.namespace = mb.m.opts.NamespaceFormat
	}

	return &mb.m, nil
}

// Returns a SMetrics instance that doesn't do anything. Function always succeeds.
func NewEmpty() *SMetrics {
	m, _ := NewSMetricBuilder(MetricsOptions{}, sink.MetricsSinkEmpty{}).Build() // we know error will always be `nil`
	return m
}

func (m *SMetrics) WithNamespacePrefix(prefix string) *SMetrics {
	shallowCopy := *m
	shallowCopy.opts.namespace += prefix
	return &shallowCopy
}

func (m *SMetrics) Count(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportCount(m.opts.namespace+name, value)
	}
}

func (m *SMetrics) Increment(name string) {
	m.sink.ReportCount(m.opts.namespace+name, 1.0)
}

func (m *SMetrics) Decrement(name string) {
	m.sink.ReportCount(m.opts.namespace+name, -1.0)
}

func (m *SMetrics) Value(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportValue(m.opts.namespace+name, value)
	}
}

func (m *SMetrics) Distribution(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportDistribution(m.opts.namespace+name, value)
	}
}

// Executes the function while measuring the time (ms) it takes to be executed
// Example usage:
// ```
// tStart := time.Now()
// // your time-sensitive logic
// metric.TimeSince("name", tStart)
// ```
// or at the start of your function you could:
// ```
// tStart := time.Now()
// defer metric.TimeSince("name", tStart)
// ```
func (m *SMetrics) TimeSince(name string, startTime time.Time) {
	m.sink.ReportDistribution(m.opts.namespace+name, time.Since(startTime).Seconds()*1000)
}

// Automatically tracks the value of a variable
func (m *SMetrics) TrackVarInt(name string, variable *int) TrackingMetric {
	ctx, ctxCancelFunc := context.WithCancel(m.ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.opts.TrackVarsPeriod):
				m.Value(name, float64(*variable))
			}
		}
	}()

	return TrackingMetric(ctxCancelFunc)
}

// Automatically tracks the value of a variable
func (m *SMetrics) TrackVarFloat(name string, variable *float64) TrackingMetric {
	ctx, ctxCancelFunc := context.WithCancel(m.ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.opts.TrackVarsPeriod):
				m.Value(name, *variable)
			}
		}
	}()

	return TrackingMetric(ctxCancelFunc)
}

// Automatically tracks the result of a function
func (m *SMetrics) TrackFuncInt(name string, f func() int) TrackingMetric {
	ctx, ctxCancelFunc := context.WithCancel(m.ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.opts.TrackVarsPeriod):
				m.Value(name, float64(f()))
			}
		}
	}()

	return TrackingMetric(ctxCancelFunc)
}

// Automatically tracks the result of a function
func (m *SMetrics) TrackFuncFloat(name string, f func() float64) TrackingMetric {
	ctx, ctxCancelFunc := context.WithCancel(m.ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.opts.TrackVarsPeriod):
				m.Value(name, f())
			}
		}
	}()

	return TrackingMetric(ctxCancelFunc)
}

// Stops tracking all the variables
func (m *SMetrics) StopTrackingVars() {
	m.ctxCancelFunc()
}
