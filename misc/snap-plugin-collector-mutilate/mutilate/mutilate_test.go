package mutilate

import (
	"fmt"
	"strings"
	"testing"
	"time"

	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMutilatePlugin(t *testing.T) {
	Convey("When I create mutilate plugin object", t, func() {
		now := time.Now()
		mutilatePlugin := NewMutilate(now)

		Convey("I should receive information about required configuration", func() {
			policy, error := mutilatePlugin.GetConfigPolicy()
			fmt.Printf("%v\t%v", *policy, error)
			So(error, ShouldBeNil)

			experimentConfig := policy.Get([]string{""}).RulesAsTable()
			So(error, ShouldBeNil)
			So(experimentConfig, ShouldHaveLength, 3)
			So(experimentConfig[0].Required, ShouldBeTrue)
			So(experimentConfig[1].Required, ShouldBeTrue)
			So(experimentConfig[2].Required, ShouldBeTrue)
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
			So(metricTypes, ShouldHaveLength, 9)
			soValidMetricType(metricTypes[0], "/intel/swan/mutilate/*/avg", "ns")
			soValidMetricType(metricTypes[1], "/intel/swan/mutilate/*/std", "ns")
			soValidMetricType(metricTypes[2], "/intel/swan/mutilate/*/min", "ns")
			soValidMetricType(metricTypes[3],
				"/intel/swan/mutilate/*/percentile/5th", "ns")
			soValidMetricType(metricTypes[4],
				"/intel/swan/mutilate/*/percentile/10th", "ns")
			soValidMetricType(metricTypes[5],
				"/intel/swan/mutilate/*/percentile/90th", "ns")
			soValidMetricType(metricTypes[6],
				"/intel/swan/mutilate/*/percentile/95th", "ns")
			soValidMetricType(metricTypes[7],
				"/intel/swan/mutilate/*/percentile/99th", "ns")
			soValidMetricType(metricTypes[8],
				"/intel/swan/mutilate/*/percentile/99_999th", "ns")

		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			configuration := cdata.NewNode()
			configuration.AddItem("stdout_file", ctypes.ConfigValueStr{
				Value: "mutilate.stdout"})
			configuration.AddItem("phase_name", ctypes.ConfigValueStr{
				Value: "this is phase name"})
			configuration.AddItem("experiment_name", ctypes.ConfigValueStr{
				Value: "this is experiment name"})
			metricTypes[0].Config_ = configuration

			metrics, error := mutilatePlugin.CollectMetrics(metricTypes)

			So(error, ShouldBeNil)
			So(metrics, ShouldHaveLength, 9)
			soValidMetric(metrics[0], "/avg", 20.8, now)
			soValidMetric(metrics[1], "/std", 23.1, now)
			soValidMetric(metrics[2], "/min", 11.9, now)
			soValidMetric(metrics[3], "/percentile/5th", 13.3, now)
			soValidMetric(metrics[4], "/percentile/10th", 13.4, now)
			soValidMetric(metrics[5], "/percentile/90th", 33.4, now)
			soValidMetric(metrics[6], "/percentile/95th", 43.1, now)
			soValidMetric(metrics[7], "/percentile/99th", 59.5, now)
			soValidMetric(metrics[8], "/percentile/99_999th", 1777.887805, now)
		})

		Convey("I should receive no metrics and error when no file path is set", func() {
			configuration := cdata.NewNode()
			configuration.AddItem("phase_name", ctypes.ConfigValueStr{
				Value: "some phase name"})
			configuration.AddItem("experiment_name", ctypes.ConfigValueStr{
				Value: "some experiment name"})
			metricTypes[0].Config_ = configuration

			metrics, error := mutilatePlugin.CollectMetrics(metricTypes)

			So(metrics, ShouldHaveLength, 0)
			So(error.Error(), ShouldEqual,
				"No file path set - no metrics are collected")

		})

		Convey("I should receive no metrics and error when mutilate results parsing fails",
			func() {
				configuration := cdata.NewNode()
				configuration.AddItem("stdout_file", ctypes.ConfigValueStr{
					Value: "mutilate_incorrect_count_of_columns.stdout"})
				configuration.AddItem("phase_name",
					ctypes.ConfigValueStr{Value: "this is phase name"})
				configuration.AddItem("experiment_name",
					ctypes.ConfigValueStr{Value: "this is experiment name"})
				metricTypes[0].Config_ = configuration

				metrics, error := mutilatePlugin.CollectMetrics(metricTypes)

				So(metrics, ShouldHaveLength, 0)
				So(error, ShouldNotBeNil)
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
