package smetrics

import (
	"math"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/luismfonseca/smetrics/sink"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// A disgraceful MetricsSink, don't use this in production
type MetricsSinkFailure struct{}

func (mse MetricsSinkFailure) Init() error {
	return errors.New("always a fail")
}

func (mse MetricsSinkFailure) ReportCount(name string, value float64)        {}
func (mse MetricsSinkFailure) ReportValue(name string, value float64)        {}
func (mse MetricsSinkFailure) ReportDistribution(name string, value float64) {}

// stores the last value it received
type MetricsSinkStoreLast struct {
	data map[string]float64
}

func (mssl *MetricsSinkStoreLast) Init() error {
	mssl.data = make(map[string]float64)
	return nil
}

func (mssl MetricsSinkStoreLast) ReportCount(name string, value float64) {
	mssl.data["count"] = value
}

func (mssl MetricsSinkStoreLast) ReportValue(name string, value float64) {
	mssl.data["value"] = value
}

func (mssl MetricsSinkStoreLast) ReportDistribution(name string, value float64) {
	mssl.data["distr"] = value
}

func (mssl MetricsSinkStoreLast) GetLastCount() float64        { return mssl.data["count"] }
func (mssl MetricsSinkStoreLast) GetLastValue() float64        { return mssl.data["value"] }
func (mssl MetricsSinkStoreLast) GetLastDistribution() float64 { return mssl.data["distr"] }

// Mocks the calls to MetricsSink
type MetricsSinkMock struct {
	mock.Mock
}

func (msm MetricsSinkMock) Init() error {
	return nil
}

func (msm *MetricsSinkMock) OnReportCount(name string, value float64) *mock.Call {
	return msm.On("ReportCount", name, value)
}

func (msm MetricsSinkMock) ReportCount(name string, value float64) {
	msm.Called(name, value)
}

func (msm *MetricsSinkMock) OnReportValue(name string, value float64) *mock.Call {
	return msm.On("ReportValue", name, value)
}

func (msm MetricsSinkMock) ReportValue(name string, value float64) {
	msm.Called(name, value)
}

func (msm *MetricsSinkMock) OnReportDistribution(name string, value float64) *mock.Call {
	return msm.On("ReportDistribution", name, value)
}

func (msm MetricsSinkMock) ReportDistribution(name string, value float64) {
	msm.Called(name, value)
}

func TestMetricsBuilder(t *testing.T) {
	Convey("A SMetricsBuilder", t, func() {
		Convey("should provide a way to instance it", func() {
			mb := NewSMetricBuilder(MetricsOptions{}, sink.MetricsSinkEmpty{})

			So(mb, ShouldNotBeNil)
		})

		Convey("should build SMetrics", func() {
			mb := NewSMetricBuilder(MetricsOptions{}, sink.MetricsSinkEmpty{})

			m, err := mb.Build()
			So(err, ShouldBeNil)
			So(m, ShouldNotBeNil)
			So(m.ctx, ShouldNotBeNil)
		})

		Convey("should fail to build SMetrics if it can't Init()", func() {
			mb := NewSMetricBuilder(MetricsOptions{}, MetricsSinkFailure{})

			m, err := mb.Build()
			So(err, ShouldNotBeNil)
			So(m, ShouldBeNil)
		})
	})
}

func TestMetrics(t *testing.T) {
	Convey("A SMetrics", t, func() {
		msMock := MetricsSinkMock{}
		m, buildErr := NewSMetricBuilder(MetricsOptions{TrackVarsPeriod: 100 * time.Millisecond}, &msMock).Build()
		So(buildErr, ShouldBeNil)
		So(m, ShouldNotBeNil)

		Convey("should forward the counts to the MetricSink", func() {
			msMock.OnReportCount("something", 123).Return().Once()
			m.Count("something", 123)

			Convey("except for NaNs", func() {
				// deliberately not setting up the mock expectation
				m.Count("something", math.NaN())
			})
		})

		Convey("should forward the values to the MetricSink", func() {
			msMock.OnReportValue("value", 234).Return().Once()
			m.Value("value", 234)

			Convey("except for NaNs", func() {
				// deliberately not setting up the mock expectation
				m.Value("value", math.NaN())
			})
		})

		Convey("should forward the distribution value to the MetricSink", func() {
			msMock.OnReportDistribution("dist", 345).Return().Once()
			m.Distribution("dist", 345)

			Convey("except for NaNs", func() {
				// deliberately not setting up the mock expectation
				m.Value("dist", math.NaN())
			})
		})

		Convey("should offer an Increment and Decrement that gets forwarded to the MetricSink", func() {
			msMock.OnReportCount("something", 1).Return().Once()
			msMock.OnReportCount("something", -1).Return().Once()
			m.Increment("something")
			m.Decrement("something")
		})

		Convey("should measure the time a function took to execute and forward it to MetricSink", func() {
			mssl := MetricsSinkStoreLast{}
			m2, err := NewSMetricBuilder(MetricsOptions{}, &mssl).Build()
			So(err, ShouldBeNil)
			So(m2, ShouldNotBeNil)

			// example usage:
			f := func() {
				tStart := time.Now()
				defer m2.TimeSince("myfunc", tStart)
				<-time.After(100 * time.Millisecond)
			}

			f()

			So(mssl.GetLastDistribution(), ShouldBeBetween, 100, 110)
		})

		Convey("should keep track of a var", func() {
			myInt := 1
			myFloat := 1.2
			msMock.OnReportValue("myInt", 1).Return().Once()
			msMock.OnReportValue("myInt", 2).Return().Once()
			msMock.OnReportValue("myFloat", 1.2).Return().Once()
			msMock.OnReportValue("myFloat", 7.3).Return().Once()

			intCancelFunc := m.TrackVarInt("myInt", &myInt)
			floatTracking := m.TrackVarFloat("myFloat", &myFloat)
			<-time.After(50 * time.Millisecond) // give it a head-start

			<-time.After(m.opts.TrackVarsPeriod)
			myInt = 2
			myFloat = 7.3
			<-time.After(m.opts.TrackVarsPeriod)

			Convey("and it should be cancelable", func() {
				intCancelFunc()
				floatTracking.Stop()
				<-time.After(m.opts.TrackVarsPeriod)
				// The mock would cause an exception if there was an unexpected call
			})
		})

		Convey("should keep track of a function result", func() {
			myInt := 1
			myFloat := 1.2
			msMock.OnReportValue("myInt", 1).Return().Once()
			msMock.OnReportValue("myInt", 2).Return().Once()
			msMock.OnReportValue("myFloat", 1.2).Return().Once()
			msMock.OnReportValue("myFloat", 7.3).Return().Once()

			intTracking := m.TrackFuncInt("myInt", func() int { return myInt })
			floatTracking := m.TrackFuncFloat("myFloat", func() float64 { return myFloat })
			<-time.After(50 * time.Millisecond) // give it a head-start

			<-time.After(m.opts.TrackVarsPeriod)
			myInt = 2
			myFloat = 7.3
			<-time.After(m.opts.TrackVarsPeriod)

			Convey("and it should be cancelable", func() {
				intTracking.Stop()
				floatTracking.Stop()
				<-time.After(m.opts.TrackVarsPeriod)
				// The mock would cause an exception if there was an unexpected call
			})
		})

		Convey("should stop keeping track of vars", func() {
			myInt := 1
			myFloat := 1.2
			m.TrackVarInt("myInt", &myInt)
			m.TrackVarFloat("myFloat", &myFloat)

			<-time.After(50 * time.Millisecond) // give it a head-start

			m.StopTrackingVars()

			<-time.After(m.opts.TrackVarsPeriod)
			// it would panic here if it tried to report as the mock isn't expecting any calls
		})

		Convey("should emit metrics namespaced", func() {
			msMock := MetricsSinkMock{}
			m, buildErr := NewSMetricBuilder(MetricsOptions{NamespaceFormat: "test."}, &msMock).Build()
			So(buildErr, ShouldBeNil)
			So(m, ShouldNotBeNil)

			msMock.OnReportCount("test.something", 123).Return().Once()
			m.Count("something", 123)

			msMock.OnReportValue("test.value", 234).Return().Once()
			m.Value("value", 234)

			msMock.OnReportDistribution("test.dist", 345).Return().Once()
			m.Distribution("dist", 345)
		})

		Convey("should allow the creation of new metrics namespaced with a prefix", func() {
			msMock := MetricsSinkMock{}
			m, buildErr := NewSMetricBuilder(MetricsOptions{NamespaceFormat: "test."}, &msMock).Build()
			So(buildErr, ShouldBeNil)
			So(m, ShouldNotBeNil)
			m2 := m.WithNamespacePrefix("new-prefix.")
			So(m2, ShouldNotBeNil)

			msMock.OnReportCount("test.something", 123).Return().Once()
			m.Count("something", 123)

			msMock.OnReportCount("test.new-prefix.something", 123).Return().Once()
			m2.Count("something", 123)
		})
	})
}
