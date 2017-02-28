package mutilate

import (
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMutilatePlugin(t *testing.T) {
	const expectedMetricsCount = 9

	Convey("When I create mutilate collector object", t, func() {
		now := time.Now()
		mutilatePlugin := NewMutilate(now)
		metricTypes, metricTypesError := mutilatePlugin.GetMetricTypes(plugin.Config{})

		Convey("I should receive information about metrics", func() {
			So(metricTypesError, ShouldBeNil)
			So(metricTypes, ShouldHaveLength, expectedMetricsCount)
			soValidMetricType(metricTypes[0], "/intel/swan/mutilate/*/avg", "ns")
			soValidMetricType(metricTypes[1], "/intel/swan/mutilate/*/std", "ns")
			soValidMetricType(metricTypes[2], "/intel/swan/mutilate/*/min", "ns")
			soValidMetricType(metricTypes[3], "/intel/swan/mutilate/*/percentile/5th", "ns")
			soValidMetricType(metricTypes[4], "/intel/swan/mutilate/*/percentile/10th", "ns")
			soValidMetricType(metricTypes[5], "/intel/swan/mutilate/*/percentile/90th", "ns")
			soValidMetricType(metricTypes[6], "/intel/swan/mutilate/*/percentile/95th", "ns")
			soValidMetricType(metricTypes[7], "/intel/swan/mutilate/*/percentile/99th", "ns")
			soValidMetricType(metricTypes[8], "/intel/swan/mutilate/*/qps", "ns")
		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			So(metricTypesError, ShouldBeNil)
			configuration := plugin.Config{}
			configuration["stdout_file"] = "mutilate.stdout"
			configuration["phase_name"] = "this is phase name"
			configuration["experiment_name"] = "this is experiment name"
			metricTypes[0].Config = configuration

			metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

			So(err, ShouldBeNil)
			So(metrics, ShouldHaveLength, expectedMetricsCount)

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
				{"/qps", 4993.1, now},
			}

			var namespace string
			for i := range metrics {
				containsMetric := false
				for _, expectedMetric := range expectedMetricsValues {
					namespace = strings.Join(append([]string{""}, metrics[i].Namespace.Strings()...), "/")
					if strings.Contains(namespace, expectedMetric.namespace) {
						soValidMetric(metrics[i], expectedMetric.namespace, expectedMetric.value, expectedMetric.date)
						containsMetric = true
						break
					}
				}
				if !containsMetric {
					t.Errorf("Expected metrics do not contain given metric %s", namespace)
				}
			}
		})

		Convey("I should receive no metrics and error when no file path is set", func() {
			So(metricTypesError, ShouldBeNil)
			configuration := plugin.Config{}
			configuration["phase_name"] = "this is phase name"
			configuration["experiment_name"] = "this is experiment name"
			metricTypes[0].Config = configuration

			metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

			So(metrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "No file path set - no metrics are collected")
		})

		Convey("I should receive no metrics and error when mutilate results parsing fails",
			func() {
				So(metricTypesError, ShouldBeNil)
				configuration := plugin.Config{}
				configuration["stdout_file"] = "mutilate_incorrect_count_of_columns.stdout"
				configuration["phase_name"] = "this is phase name"
				configuration["experiment_name"] = "this is experiment name"
				metricTypes[0].Config = configuration

				metrics, err := mutilatePlugin.CollectMetrics(metricTypes)

				So(metrics, ShouldHaveLength, 0)
				So(err, ShouldNotBeNil)
			})
	})
}

func soValidMetricType(metricType plugin.Metric, namespace string, unit string) {
	So(strings.Join(append([]string{""}, metricType.Namespace.Strings()...), "/"), ShouldEqual, namespace)
	So(metricType.Unit, ShouldEqual, unit)
	So(metricType.Version, ShouldEqual, 1)
}

func soValidMetric(metric plugin.Metric, namespaceSuffix string, value float64, time time.Time) {
	namespace := strings.Join(append([]string{""}, metric.Namespace.Strings()...), "/")
	So(namespace, ShouldStartWith, "/intel/swan/mutilate/")
	So(namespace, ShouldEndWith, namespaceSuffix)
	So(strings.Contains(namespace, "*"), ShouldBeFalse)
	So(metric.Unit, ShouldEqual, "ns")
	So(metric.Tags, ShouldHaveLength, 0)
	data, typeFound := metric.Data.(float64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
	So(metric.Timestamp.Unix(), ShouldEqual, time.Unix())
}
