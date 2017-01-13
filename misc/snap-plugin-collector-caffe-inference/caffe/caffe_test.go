package caffe

import (
	"strings"
	"testing"

	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

type metric struct {
	namespace string
	value     uint64
}

var (
	expectedMetricsCount = 1
	expectedMetric       = metric{"/batches", 0}
)

func TestCaffeInferenceCollectorPlugin(t *testing.T) {
	Convey("When I create caffe-inference plugin object", t, func() {
		caffePlugin := InferenceCollector{}

		Convey("I should receive meta data for plugin", func() {
			meta := Meta()
			So(meta.Name, ShouldEqual, "caffe-inference")
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
			soValidMetricType(metricTypes[0], "/intel/swan/caffe/inference/*/batches", METRICNAME)
		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			configuration := makeDefaultConfiguration("log-finished.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
			So(collectedMetrics[0].Namespace().String(), ShouldContainSubstring, expectedMetric.namespace)
			expectedMetric.value = 99
			soValidMetric(collectedMetrics[0], expectedMetric)

		})
		Convey("I should receive no metrics end error when caffe ended prematurely ", func() {
			configuration := makeDefaultConfiguration("log-notstarted.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(collectedMetrics, ShouldHaveLength, 0)
			So(err, ShouldEqual, ErrParse)
		})
		Convey("I should receive no metrics end error when caffe ended prematurely and there is single work 'Batch' in log without number", func() {
			configuration := makeDefaultConfiguration("log-interrupted2.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(collectedMetrics, ShouldHaveLength, 0)
			So(err, ShouldEqual, ErrParse)
		})
		Convey("I should receive valid metric when caffe was killed during inference", func() {
			configuration := makeDefaultConfiguration("log-interrupted.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
			So(collectedMetrics[0].Namespace().String(), ShouldContainSubstring, expectedMetric.namespace)
			expectedMetric.value = 24
			soValidMetric(collectedMetrics[0], expectedMetric)
		})
		Convey("I should receive valid metric when caffe was killed during inference and last word in log is 'Batch'", func() {
			configuration := makeDefaultConfiguration("log-interrupted3.txt")
			metricTypes[0].Config_ = configuration
			collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
			So(collectedMetrics[0].Namespace().String(), ShouldContainSubstring, expectedMetric.namespace)
			expectedMetric.value = 3
			soValidMetric(collectedMetrics[0], expectedMetric)
		})
		Convey("I should receive no metrics and error when no file path is set", func() {
			configuration := cdata.NewNode()
			metricTypes[0].Config_ = configuration
			metrics, err := caffePlugin.CollectMetrics(metricTypes)
			So(metrics, ShouldHaveLength, 0)
			So(err, ShouldEqual, ErrConf)
		})
	})
}

func makeDefaultConfiguration(fileName string) *cdata.ConfigDataNode {
	configuration := cdata.NewNode()
	configuration.AddItem("stdout_file", ctypes.ConfigValueStr{Value: fileName})
	return configuration
}

func soValidMetricType(metricType snapPlugin.MetricType, namespace string, unit string) {
	So(metricType.Namespace().String(), ShouldEqual, namespace)
	So(metricType.Unit(), ShouldEqual, unit)
	So(metricType.Version(), ShouldEqual, 1)
}

func soValidMetric(metric snapPlugin.MetricType, expectedMetric metric) {
	namespaceSuffix := expectedMetric.namespace
	namespacePrefix := strings.Join(append([]string{""}, namespace...), "/")
	value := expectedMetric.value

	So(metric.Namespace().String(), ShouldStartWith, namespacePrefix)
	So(metric.Namespace().String(), ShouldEndWith, namespaceSuffix)
	So(strings.Contains(metric.Namespace().String(), "*"), ShouldBeFalse)
	So(metric.Unit(), ShouldEqual, METRICNAME)
	So(metric.Tags(), ShouldHaveLength, 0)
	data, typeFound := metric.Data().(uint64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
}
