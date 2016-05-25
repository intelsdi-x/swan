package mutilate

import (
	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
	"time"
)

func TestMutilatePlugin(t *testing.T) {
	Convey("When I create mutilate plugin object", t, func() {
		now := time.Now()
		mutilatePlugin := NewMutilate(now)

		Convey("I should receive information about required configuration", func() {
			policy, err := mutilatePlugin.GetConfigPolicy()
			So(err, ShouldBeNil)

			experimentConfig := policy.Get([]string{""}).RulesAsTable()
			So(err, ShouldBeNil)
			So(experimentConfig, ShouldHaveLength, 1)
			So(experimentConfig[0].Required, ShouldBeTrue)
			So(experimentConfig[0].Name, ShouldEqual, "stdout_file")
			So(experimentConfig[0].Type, ShouldEqual, "string")
		})

		config := snapPlugin.NewPluginConfigType()
		phaseName := ctypes.ConfigValueStr{Value: "some random tag!"}
		config.AddItem("phase_name", phaseName)
		config.AddItem("experiment_name",
			ctypes.ConfigValueStr{Value: "some random experiment!"})
		config.AddItem("stdout_file", ctypes.ConfigValueStr{Value: "mutilate.stdout"})

		metricTypes, metricTypesError := mutilatePlugin.GetMetricTypes(config)

		Convey("I should receive information about metrics", func() {
			So(metricTypesError, ShouldBeNil)
			So(metricTypes, ShouldHaveLength, 11)
			soValidMetricType(metricTypes[0], "/intel/swan/mutilate/*/avg", "ns")
			soValidMetricType(metricTypes[1], "/intel/swan/mutilate/*/std", "ns")
			soValidMetricType(metricTypes[2], "/intel/swan/mutilate/*/min", "ns")
			soValidMetricType(metricTypes[3], "/intel/swan/mutilate/*/percentile/5th", "ns")
			soValidMetricType(metricTypes[4], "/intel/swan/mutilate/*/percentile/10th", "ns")
			soValidMetricType(metricTypes[5], "/intel/swan/mutilate/*/percentile/90th", "ns")
			soValidMetricType(metricTypes[6], "/intel/swan/mutilate/*/percentile/95th", "ns")
			soValidMetricType(metricTypes[7], "/intel/swan/mutilate/*/percentile/99th", "ns")
			soValidMetricType(metricTypes[8], "/intel/swan/mutilate/*/qps/total", "ns")
			soValidMetricType(metricTypes[9], "/intel/swan/mutilate/*/qps/peak", "ns")
			soValidMetricType(metricTypes[10], "/intel/swan/mutilate/*/percentile/*/custom", "ns")

		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			So(metricTypesError, ShouldBeNil)
			configuration := cdata.NewNode()
			configuration.AddItem("stdout_file", ctypes.ConfigValueStr{
				Value: "mutilate.stdout"})
			configuration.AddItem("phase_name", ctypes.ConfigValueStr{
				Value: "this is phase name"})
			configuration.AddItem("experiment_name", ctypes.ConfigValueStr{
				Value: "this is experiment name"})
			metricTypes[0].Config_ = configuration

			metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, 11)

			type metric struct {
				namespace string
				value     float64
				date      time.Time
			}

			var expectedMetricsValues = []metric{
				{"/avg", 20.8, now},
				{"/std", 23.1, now},
				{"/min", 11.9, now},
				{"/percentile/5th", 13.3, now},
				{"/percentile/10th", 13.4, now},
				{"/percentile/90th", 33.4, now},
				{"/percentile/95th", 43.1, now},
				{"/percentile/99th", 59.5, now},
				{"/percentile/99_999th/custom", 1777.887805, now},
			}

			for i := range metrics {
				containsMetric := false
				for _, expectedMetric := range expectedMetricsValues {
					if strings.Contains(metrics[i].Namespace().String(), expectedMetric.namespace) {
						soValidMetric(metrics[i], expectedMetric.namespace, expectedMetric.value,
							expectedMetric.date)
						containsMetric = true
						break
					}
				}
				if !containsMetric {
					t.Error("Expected metrics do not contain given metric " +
						metrics[i].Namespace().String())
				}
			}

			So(metrics[8].Namespace().String(), ShouldNotEndWith, "percentile/percentile/99_999th/custom")
		})

		Convey("I should receive no metrics and error when no file path is set", func() {
			So(metricTypesError, ShouldBeNil)
			configuration := cdata.NewNode()
			configuration.AddItem("phase_name", ctypes.ConfigValueStr{
				Value: "some phase name"})
			configuration.AddItem("experiment_name", ctypes.ConfigValueStr{
				Value: "some experiment name"})
			metricTypes[0].Config_ = configuration

			metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

			So(metrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "No file path set - no metrics are collected")
		})

		Convey("I should receive no metrics and error when mutilate results parsing fails",
			func() {
				So(metricTypesError, ShouldBeNil)
				configuration := cdata.NewNode()
				configuration.AddItem("stdout_file", ctypes.ConfigValueStr{
					Value: "mutilate_incorrect_count_of_columns.stdout"})
				configuration.AddItem("phase_name",
					ctypes.ConfigValueStr{Value: "this is phase name"})
				configuration.AddItem("experiment_name",
					ctypes.ConfigValueStr{Value: "this is experiment name"})
				metricTypes[0].Config_ = configuration

				metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

				So(metrics, ShouldHaveLength, 0)
				So(err, ShouldNotBeNil)
			})
	})
}

func soValidMetricType(metricType snapPlugin.MetricType, namespace string, unit string) {
	So(metricType.Namespace().String(), ShouldEqual, namespace)
	So(metricType.Unit(), ShouldEqual, unit)
	So(metricType.Version(), ShouldEqual, 1)
}

func soValidMetric(metric snapPlugin.MetricType, namespaceSuffix string, value float64, time time.Time) {
	So(metric.Namespace().String(), ShouldStartWith, "/intel/swan/mutilate/")
	So(metric.Namespace().String(), ShouldEndWith, namespaceSuffix)
	So(strings.Contains(metric.Namespace().String(), "*"), ShouldBeFalse)
	So(metric.Unit(), ShouldEqual, "ns")
	So(metric.Tags(), ShouldHaveLength, 0)
	So(metric.Data().(float64), ShouldEqual, value)
	So(metric.Timestamp().Unix(), ShouldEqual, time.Unix())
}
