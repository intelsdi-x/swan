package caffe

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
	expectedMetricsCount = 1
	now                  = time.Now()
	expectedMetrics      = []metric{{"/img", 0, now}}
)

func TestCaffeInferenceCollectorPlugin(t *testing.T) {
	Convey("When I create Caffe-inference plugin object", t, func() {
		caffePlugin := InferenceCollector{}

		Convey("I should receive meta data for plugin", func() {
			meta := Meta()
			So(meta.Name, ShouldEqual, "caffeinference")
			So(meta.Version, ShouldEqual, 1)
			So(meta.Type, ShouldEqual, snapPlugin.CollectorPluginType)
		})

		Convey("I should receive information about required configuration", func() {
			policy, err := caffePlugin.GetConfigPolicy()
			So(err, ShouldBeNil)

			experimentConfig := policy.Get(namespace).RulesAsTable()
			So(experimentConfig, ShouldHaveLength, 1)
			So(experimentConfig[0].Required, ShouldBeTrue)
			So(experimentConfig[0].Name, ShouldEqual, "stdout_file")
			So(experimentConfig[0].Type, ShouldEqual, "string")
		})

		config := snapPlugin.NewPluginConfigType()
		metricTypes, err := caffePlugin.GetMetricTypes(config)
		So(err, ShouldBeNil)

		Convey("I should receive information about metrics", func() {
			So(metricTypes, ShouldHaveLength, expectedMetricsCount)
			soValidMetricType(metricTypes[0], "/intel/swan/caffe/inference/*/img", "images")
		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			configuration := makeDefaultConfiguration("log-finished.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)

			// Check whether expected metrics contain the metric.
			var found bool
			for _, metric := range collectedMetrics {
				found = false
				for _, expectedMetric := range expectedMetrics {
					if strings.Contains(metric.Namespace().String(), expectedMetric.namespace) {
						expectedMetric.value = 990000
						soValidMetric(metric, expectedMetric)
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected metrics do not contain the metric %s", metric.Namespace().String())
				}
			}
		})
		Convey("I should receive no metrics end error when caffe ended prematurely ", func() {
			configuration := makeDefaultConfiguration("log-notstarted.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(collectedMetrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "Did not find batch number in the output log")
		})
		Convey("I should receive no metrics end error when caffe ended prematurely and there is single work 'Batch' in log without number", func() {
			configuration := makeDefaultConfiguration("log-interrupted2.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(collectedMetrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "Did not find batch number in the output log")
		})
		Convey("I should receive valid metric when caffe was killed during inference", func() {
			configuration := makeDefaultConfiguration("log-interrupted.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)

			// Check whether expected metrics contain the metric.
			var found bool
			for _, metric := range collectedMetrics {
				found = false
				for _, expectedMetric := range expectedMetrics {
					if strings.Contains(metric.Namespace().String(), expectedMetric.namespace) {
						expectedMetric.value = 240000
						soValidMetric(metric, expectedMetric)
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected metrics do not contain the metric %s", metric.Namespace().String())
				}
			}

		})
		Convey("I should receive valid metric when caffe was killed during inference and last word in log is 'Batch'", func() {
			configuration := makeDefaultConfiguration("log-interrupted3.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)

			// Check whether expected metrics contain the metric.
			var found bool
			for _, metric := range collectedMetrics {
				found = false
				for _, expectedMetric := range expectedMetrics {
					if strings.Contains(metric.Namespace().String(), expectedMetric.namespace) {
						expectedMetric.value = 30000
						soValidMetric(metric, expectedMetric)
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected metrics do not contain the metric %s", metric.Namespace().String())
				}
			}

		})
		Convey("I should receive no metrics and error when no file path is set", func() {
			configuration := cdata.NewNode()
			metricTypes[0].Config_ = configuration
			metrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(metrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "No file path set - no metrics are going to be collected")
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

func soValidMetric(metric snapPlugin.MetricType, expectedMetric metric) {
	namespaceSuffix := expectedMetric.namespace
	value := expectedMetric.value
	time := expectedMetric.date

	So(metric.Namespace().String(), ShouldStartWith, "/intel/swan/caffe/inference/")
	So(metric.Namespace().String(), ShouldEndWith, namespaceSuffix)
	So(strings.Contains(metric.Namespace().String(), "*"), ShouldBeFalse)
	So(metric.Unit(), ShouldEqual, "images")
	So(metric.Tags(), ShouldHaveLength, 0)
	data, typeFound := metric.Data().(uint64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
	So(metric.Timestamp().Unix(), ShouldEqual, time.Unix())
}
