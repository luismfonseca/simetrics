package sink

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/luismfonseca/simetrics/type/distribution"
	"github.com/sirupsen/logrus"
)

type MetricsSinkDogStatsD struct {
	tags          []string // `source:` + defaults to hostname
	context       context.Context
	statsDClient  *statsd.Client
	mutex         sync.Mutex
	counts        map[string]float64
	distributions map[string]*distribution.Distribution
	log           *logrus.Entry
}

func NewMetricsSinkDogStatsD(address, sourceFormat string, log *logrus.Entry) *MetricsSinkDogStatsD {
	source := sourceFormat
	if strings.Contains(sourceFormat, "%s") {
		hostname, err := os.Hostname()
		if err != nil {
			log.WithError(err).Warnln("Failed to get the hostname to use as the metric Source. Proceeding with an empty string.")
		}
		source = fmt.Sprintf(sourceFormat, hostname)
	}

	client, err := statsd.New(address)
	if err != nil {
		log.WithField("address", address).Fatalf("Could not setup statsD client")
	}

	return &MetricsSinkDogStatsD{
		tags:          []string{"source:" + source},
		context:       context.Background(),
		statsDClient:  client,
		counts:        map[string]float64{},
		distributions: map[string]*distribution.Distribution{},
		log:           log,
	}
}

func (msl *MetricsSinkDogStatsD) run() {
	ticker := time.NewTicker(SubmitPeriodSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := msl.statsDClient.Flush()
			if err != nil {
				msl.log.WithError(err).Warnln("Failed to post metrics")
			}
		case <-msl.context.Done():
			msl.log.Infoln("Terminating StatsD Sink.")
			return
		}
	}
}

func (msl *MetricsSinkDogStatsD) Init() error {
	go msl.run()

	return nil
}

func (msl *MetricsSinkDogStatsD) ReportCount(name string, value float64) {
	_ = msl.statsDClient.Count(name, int64(value), msl.tags, 1)
}

func (msl *MetricsSinkDogStatsD) ReportValue(name string, value float64) {
	_ = msl.statsDClient.Gauge(name, value, msl.tags, 1)
}

func (msl *MetricsSinkDogStatsD) ReportDistribution(name string, value float64) {
	_ = msl.statsDClient.Distribution(name, value, msl.tags, 1)
}
