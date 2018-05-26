package sink

import (
	"sync"
	"time"

	"github.com/luismfonseca/simetrics/type/distribution"
	"github.com/sirupsen/logrus"
)

const (
	StdoutFlushPeriod = 5 * time.Second
)

type MetricsSinkStdout struct {
	mutex         sync.Mutex
	counts        map[string]float64
	distributions map[string]*distribution.Distribution
	log           *logrus.Entry
}

func NewMetricsSinkStdout(log *logrus.Entry) *MetricsSinkStdout {
	return &MetricsSinkStdout{
		counts:        map[string]float64{},
		distributions: map[string]*distribution.Distribution{},
		log:           log,
	}
}

func (msl *MetricsSinkStdout) Init() error {
	go msl.run()

	return nil
}

func (msl *MetricsSinkStdout) run() {
	ticker := time.Tick(StdoutFlushPeriod)

	for {
		select {
		case <-ticker:
			msl.mutex.Lock()

			for name, value := range msl.counts {
				msl.log.WithField(name, value).Println("Metric report")
			}
			for name, dist := range msl.distributions {
				msl.log.WithField(name, dist).Println("Metric report")
			}

			msl.distributions = map[string]*distribution.Distribution{}
			msl.counts = map[string]float64{}

			msl.mutex.Unlock()
		}
	}
}

func (msl *MetricsSinkStdout) ReportCount(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	msl.counts[name] += value
}

func (msl *MetricsSinkStdout) ReportValue(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	msl.counts[name] = value
}

func (msl *MetricsSinkStdout) ReportDistribution(name string, value float64) {
	msl.mutex.Lock()
	defer msl.mutex.Unlock()

	histogram, ok := msl.distributions[name]
	if ok {
		histogram.AddEntry(value)
	} else {
		msl.distributions[name] = distribution.FromValue(value)
	}
}
