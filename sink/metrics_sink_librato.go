package sink

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/heroku/go-metrics-librato"
	"github.com/luismfonseca/smetrics/type/distribution"
	"github.com/sirupsen/logrus"
)

const (
	SubmitPeriodSeconds = 5
)

type MetricsSinkLibrato struct {
	Email     string
	Token     string
	Namespace string
	Source    string // defaults to hostname

	context       context.Context
	mutex         sync.Mutex
	counts        map[string]float64
	distributions map[string]*distribution.Distribution
	log           *logrus.Entry
}

func NewMetricsSinkLibrato(email, token, namespace, sourceFormat string, log *logrus.Entry) *MetricsSinkLibrato {
	source := sourceFormat
	if strings.Contains(sourceFormat, "%s") {
		hostname, err := os.Hostname()
		if err != nil {
			log.WithError(err).Warnln("Failed to get the hostname to use as the metric Source. Proceeding with an empty string.")
		}
		source = fmt.Sprintf(sourceFormat, hostname)
	}

	return &MetricsSinkLibrato{
		Email:     email,
		Token:     token,
		Namespace: namespace,
		Source:    source,

		context:       context.Background(),
		counts:        map[string]float64{},
		distributions: map[string]*distribution.Distribution{},
		log:           log,
	}
}

func (msl *MetricsSinkLibrato) buildBatch() librato.Batch {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	batch := librato.Batch{
		// coerce timestamps to a stepping fn so that they line up in Librato graphs
		MeasureTime: (time.Now().Unix() / SubmitPeriodSeconds) * SubmitPeriodSeconds,
		Source:      msl.Source,
		Gauges:      make([]librato.Measurement, 0),
		Counters:    make([]librato.Measurement, 0),
	}

	for name, value := range msl.counts {
		batch.Gauges = append(batch.Gauges, librato.Measurement{
			"name":  name,
			"value": value,
		})
	}
	for name, dist := range msl.distributions {
		batch.Gauges = append(batch.Gauges, librato.Measurement{
			"name":        name,
			"count":       dist.N,
			"min":         dist.Min,
			"max":         dist.Max,
			"sum":         dist.SumX,
			"sum_squares": dist.SumX2,
		})
	}
	msl.distributions = map[string]*distribution.Distribution{}
	msl.counts = map[string]float64{}

	return batch
}

func (msl *MetricsSinkLibrato) run() {
	ticker := time.Tick(SubmitPeriodSeconds * time.Second)
	metricsApi := &librato.LibratoClient{Email: msl.Email, Token: msl.Token}

	for {
		select {
		case <-ticker:
			metricsBatch := msl.buildBatch()
			msl.log.
				WithField("batch_size", len(metricsBatch.Counters)+len(metricsBatch.Gauges)).
				Debug("Posting librato metrics...")

			err := metricsApi.PostMetrics(metricsBatch)
			if err != nil {
				msl.log.WithError(err).Warnln("Failed to post metrics")
			}

		case <-msl.context.Done():
			msl.log.Infoln("Terminating Librato Sink.")
			return
		}
	}
}

func (msl *MetricsSinkLibrato) Init() error {
	go msl.run()

	return nil
}

func (msl *MetricsSinkLibrato) ReportCount(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	msl.counts[name] += value
}

func (msl *MetricsSinkLibrato) ReportValue(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	msl.counts[name] = value
}

func (msl *MetricsSinkLibrato) ReportDistribution(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	histogram, ok := msl.distributions[name]
	if ok {
		histogram.AddEntry(value)
	} else {
		msl.distributions[name] = distribution.FromValue(value)
	}
}
