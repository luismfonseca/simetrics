package simetrics

import (
	"context"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/luismfonseca/simetrics/sink"
)

type MetricsOptions struct {
	// The period in which variables should be tracked
	TrackVarsPeriod time.Duration

	// A prefix with the app name to fill the `%s`
	NamespaceFormat string

	// computed namespaceFormat + app name
	namespace string
}

type SiMetrics struct {
	sink          sink.MetricsSink
	opts          MetricsOptions
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
}

type SiMetricsBuilder struct {
	m SiMetrics
}

// A tracking metric that can be stopped.
type TrackingMetric context.CancelFunc

func (t TrackingMetric) Stop() {
	t()
}

// Builds a `SiMetricsBuilder` so you can call `Build()` to get the built `SiMetrics`
func NewBuilder(options MetricsOptions, ms sink.MetricsSink) *SiMetricsBuilder {
	if options.TrackVarsPeriod == 0 {
		options.TrackVarsPeriod = 5 * time.Second
	}

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	return &SiMetricsBuilder{m: SiMetrics{sink: ms, opts: options, ctx: ctx, ctxCancelFunc: ctxCancelFunc}}
}

// Returns a built and fully initialized `SiMetrics` or an error
func (mb *SiMetricsBuilder) Build() (*SiMetrics, error) {
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

// Returns a SiMetrics instance that doesn't do anything. Function always succeeds.
func NewEmpty() *SiMetrics {
	m, _ := NewBuilder(MetricsOptions{}, sink.MetricsSinkEmpty{}).Build() // we know error will always be `nil`
	return m
}

func (m *SiMetrics) WithNamespacePrefix(prefix string) *SiMetrics {
	shallowCopy := *m
	shallowCopy.opts.namespace += prefix
	return &shallowCopy
}

func (m *SiMetrics) Count(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportCount(m.opts.namespace+name, value)
	}
}

func (m *SiMetrics) Increment(name string) {
	m.sink.ReportCount(m.opts.namespace+name, 1.0)
}

func (m *SiMetrics) Decrement(name string) {
	m.sink.ReportCount(m.opts.namespace+name, -1.0)
}

func (m *SiMetrics) Value(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportValue(m.opts.namespace+name, value)
	}
}

func (m *SiMetrics) Distribution(name string, value float64) {
	if !math.IsNaN(value) {
		m.sink.ReportDistribution(m.opts.namespace+name, value)
	}
}

// Measures the time in ms since the given `startTime`
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
func (m *SiMetrics) TimeSince(name string, startTime time.Time) {
	m.sink.ReportDistribution(m.opts.namespace+name, time.Since(startTime).Seconds()*1000)
}

// Automatically tracks the value of a variable
func (m *SiMetrics) TrackVarInt(name string, variable *int) TrackingMetric {
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
func (m *SiMetrics) TrackVarFloat(name string, variable *float64) TrackingMetric {
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
func (m *SiMetrics) TrackFuncInt(name string, f func() int) TrackingMetric {
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
func (m *SiMetrics) TrackFuncFloat(name string, f func() float64) TrackingMetric {
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
func (m *SiMetrics) StopTrackingVars() {
	m.ctxCancelFunc()
}
