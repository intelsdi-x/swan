package specjbb

import (
	"strings"
	"testing"
	"time"

	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

type metric struct {
	namespace string
	value     uint64
	date      time.Time
}

var (
	expectedMetricsCount = 6
	now                  = time.Now()
	expectedMetrics      = []metric{
		{"/min", 300, now},
		{"/percentile/50th", 3100, now},
		{"/percentile/90th", 21000, now},
		{"/percentile/95th", 89000, now},
		{"/percentile/99th", 517000, now},
		{"/max", 640000, now}}
)

func TestSpecjbbCollectorPlugin(t *testing.T) {
	Convey("When I create SPECjbb plugin object", t, func() {
		specjbbPlugin := NewSpecjbb(now)

		Convey("Collector should not be nil", func() {
			So(specjbbPlugin, ShouldNotBeNil)
		})

		Convey("I should receive meta data for plugin", func() {
			meta := Meta()
			So(meta.Name, ShouldEqual, "specjbb")
			So(meta.Version, ShouldEqual, 1)
			So(meta.Type, ShouldEqual, snapPlugin.CollectorPluginType)
		})

		Convey("I should receive information about required configuration", func() {
			policy, err := specjbbPlugin.GetConfigPolicy()
			So(err, ShouldBeNil)

			experimentConfig := policy.Get([]string{""}).RulesAsTable()
			So(experimentConfig, ShouldHaveLength, 1)
			So(experimentConfig[0].Required, ShouldBeTrue)
			So(experimentConfig[0].Name, ShouldEqual, "stdout_file")
			So(experimentConfig[0].Type, ShouldEqual, "string")
		})

		config := snapPlugin.NewPluginConfigType()
		metricTypes, err := specjbbPlugin.GetMetricTypes(config)
		So(err, ShouldBeNil)

		Convey("I should receive information about metrics", func() {
			So(metricTypes, ShouldHaveLength, expectedMetricsCount)
			soValidMetricType(metricTypes[0], "/intel/swan/specjbb/*/min", "ns")
			soValidMetricType(metricTypes[1], "/intel/swan/specjbb/*/max", "ns")
			soValidMetricType(metricTypes[2], "/intel/swan/specjbb/*/percentile/50th", "ns")
			soValidMetricType(metricTypes[3], "/intel/swan/specjbb/*/percentile/90th", "ns")
			soValidMetricType(metricTypes[4], "/intel/swan/specjbb/*/percentile/95th", "ns")
			soValidMetricType(metricTypes[5], "/intel/swan/specjbb/*/percentile/99th", "ns")

		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			configuration := makeDefaultConfiguration("specjbb.stdout")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := specjbbPlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)

			// Check whether gathered metrics contain all expected metric. If not, return not found one.
			found, metricNamespace := containMetrics(collectedMetrics, expectedMetrics)
			if !found {
				t.Errorf("Collected metrics do not contain expected metric %s", metricNamespace)
			}
		})

		Convey("I should receive no metrics and error when no file path is set", func() {
			configuration := cdata.NewNode()
			metricTypes[0].Config_ = configuration
			metrics, err := specjbbPlugin.CollectMetrics(metricTypes)
			So(metrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "No file path set - no metrics are collected")
		})

		Convey("I should receive no metrics and error when SPECjbb results parsing fails", func() {
			configuration := makeDefaultConfiguration("specjbb_incorrect_format.stdout")
			metricTypes[0].Config_ = configuration

			metrics, err := specjbbPlugin.CollectMetrics(metricTypes)

			So(metrics, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
		})
	})
}

func makeDefaultConfiguration(fileName string) *cdata.ConfigDataNode {
	configuration := cdata.NewNode()
	configuration.AddItem("stdout_file", ctypes.ConfigValueStr{Value: fileName})
	configuration.AddItem("phase_name", ctypes.ConfigValueStr{Value: "phase name"})
	configuration.AddItem("experiment_name", ctypes.ConfigValueStr{Value: "experiment name"})
	return configuration
}

func soValidMetricType(metricType snapPlugin.MetricType, namespace string, unit string) {
	So(metricType.Namespace().String(), ShouldEqual, namespace)
	So(metricType.Unit(), ShouldEqual, unit)
	So(metricType.Version(), ShouldEqual, 1)
}

func containMetrics(collectedMetrics []snapPlugin.MetricType, expectedMetrics []metric) (contain bool, namespace string) {
	for _, metric := range collectedMetrics {
		for _, expectedMetric := range expectedMetrics {
			if strings.Contains(metric.Namespace().String(), expectedMetric.namespace) {
				soValidMetric(metric, expectedMetric)
				return true, metric.Namespace().String()
			}
		}
		return false, metric.Namespace().String()
	}
	return
}

func soValidMetric(metric snapPlugin.MetricType, expectedMetric metric) {
	namespaceSuffix := expectedMetric.namespace
	value := expectedMetric.value
	time := expectedMetric.date

	So(metric.Namespace().String(), ShouldStartWith, "/intel/swan/specjbb/")
	So(metric.Namespace().String(), ShouldEndWith, namespaceSuffix)
	So(strings.Contains(metric.Namespace().String(), "*"), ShouldBeFalse)
	So(metric.Unit(), ShouldEqual, "ns")
	So(metric.Tags(), ShouldHaveLength, 0)
	data, typeFound := metric.Data().(uint64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
	So(metric.Timestamp().Unix(), ShouldEqual, time.Unix())
}
