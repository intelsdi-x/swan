package mutilate

import (
	"github.com/intelsdi-x/snap/control/plugin"
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
			policy, error := mutilatePlugin.GetConfigPolicy()

			phaseName := policy.Get([]string{"experiment"}).RulesAsTable()
			So(error, ShouldBeNil)
			So(phaseName, ShouldHaveLength, 1)
			So(phaseName[0].Required, ShouldBeTrue)
			So(phaseName[0].Type, ShouldEqual, "string")
			So(phaseName[0].Name, ShouldEqual, "phase_name")
		})

		config := plugin.NewPluginConfigType()
		phaseName := ctypes.ConfigValueStr{Value: "some random tag!"}
		config.AddItem("phase_name", phaseName)

		metricTypes, metricTypesError := mutilatePlugin.GetMetricTypes(config)

		Convey("I should receive information about metrics", func() {
			So(metricTypesError, ShouldBeNil)
			So(metricTypes, ShouldHaveLength, 9)
			soValidMetricType(metricTypes[0], "/intel/swan/mutilate/*/avg", "ns", "some random tag!")
			soValidMetricType(metricTypes[1], "/intel/swan/mutilate/*/std", "ns", "some random tag!")
			soValidMetricType(metricTypes[2], "/intel/swan/mutilate/*/min", "ns", "some random tag!")
			soValidMetricType(metricTypes[3], "/intel/swan/mutilate/*/percentile/5th", "ns", "some random tag!")
			soValidMetricType(metricTypes[4], "/intel/swan/mutilate/*/percentile/10th", "ns", "some random tag!")
			soValidMetricType(metricTypes[5], "/intel/swan/mutilate/*/percentile/90th", "ns", "some random tag!")
			soValidMetricType(metricTypes[6], "/intel/swan/mutilate/*/percentile/95th", "ns", "some random tag!")
			soValidMetricType(metricTypes[7], "/intel/swan/mutilate/*/percentile/99th", "ns", "some random tag!")
			soValidMetricType(metricTypes[8], "/intel/swan/mutilate/*/percentile/99.999th", "ns", "some random tag!")
		})

		Convey("I should receive valid metrics when I try to collect them", func() {
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
			soValidMetric(metrics[8], "/percentile/99.999th", 1777.887805, now)
		})
	})
}

func soValidMetricType(metricType plugin.MetricType, namespace string, unit string, tag string) {
	So(metricType.Namespace().String(), ShouldEqual, namespace)
	So(metricType.Unit(), ShouldEqual, unit)
	So(metricType.Tags()["phase_name"], ShouldEqual, tag)
	So(metricType.Version(), ShouldEqual, 1)
}

func soValidMetric(metric plugin.MetricType, namespaceSuffix string, value float64, time time.Time) {
	So(metric.Namespace().String(), ShouldStartWith, "/intel/swan/mutilate/")
	So(metric.Namespace().String(), ShouldEndWith, namespaceSuffix)
	So(strings.Contains(metric.Namespace().String(), "*"), ShouldBeFalse)
	So(metric.Unit(), ShouldEqual, "ns")
	So(metric.Tags()["phase_name"], ShouldEqual, "some random tag!")
	So(metric.Data().(float64), ShouldEqual, value)
	So(metric.Timestamp().Unix(), ShouldEqual, time.Unix())
}
