// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package specjbb

import (
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

type metric struct {
	namespace string
	value     uint64
	date      time.Time
}

var (
	expectedMetricsCount = 8
	now                  = time.Now()
	expectedMetrics      = []metric{
		{"/min", 300, now},
		{"/percentile/50th", 3100, now},
		{"/percentile/90th", 21000, now},
		{"/percentile/95th", 89000, now},
		{"/percentile/99th", 517000, now},
		{"/max", 640000, now},
		{"/qps", 4007, now},
		{"/issued_requests", 4007, now}}
)

func TestSpecjbbCollectorPlugin(t *testing.T) {
	Convey("When I create SPECjbb plugin object", t, func() {
		specjbbPlugin := NewSpecjbb(now)
		metricTypes, err := specjbbPlugin.GetMetricTypes(plugin.Config{})
		So(err, ShouldBeNil)

		Convey("I should receive information about metrics", func() {
			So(metricTypes, ShouldHaveLength, expectedMetricsCount)
			soValidMetricType(metricTypes[0], "/intel/swan/specjbb/*/min", "ns")
			soValidMetricType(metricTypes[1], "/intel/swan/specjbb/*/max", "ns")
			soValidMetricType(metricTypes[2], "/intel/swan/specjbb/*/percentile/50th", "ns")
			soValidMetricType(metricTypes[3], "/intel/swan/specjbb/*/percentile/90th", "ns")
			soValidMetricType(metricTypes[4], "/intel/swan/specjbb/*/percentile/95th", "ns")
			soValidMetricType(metricTypes[5], "/intel/swan/specjbb/*/percentile/99th", "ns")
			soValidMetricType(metricTypes[6], "/intel/swan/specjbb/*/qps", "ns")
			soValidMetricType(metricTypes[7], "/intel/swan/specjbb/*/issued_requests", "ns")

		})

		Convey("I should receive valid metrics when I try to collect them", func() {
			configuration := makeDefaultConfiguration("specjbb.stdout")
			metricTypes[0].Config = configuration
			collectedMetrics, err := specjbbPlugin.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)

			// Check whether expected metrics contain the metric.
			var found bool
			for _, metric := range collectedMetrics {
				var namespace string
				found = false
				for _, expectedMetric := range expectedMetrics {
					namespace = "/" + strings.Join(metric.Namespace.Strings(), "/")
					if strings.Contains(namespace, expectedMetric.namespace) {
						soValidMetric(metric, expectedMetric)
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected metrics do not contain the metric %s", namespace)
				}
			}
		})

		Convey("I should receive no metrics and error when no file path is set", func() {
			configuration := plugin.Config{}
			metricTypes[0].Config = configuration
			metrics, err := specjbbPlugin.CollectMetrics(metricTypes)
			So(metrics, ShouldHaveLength, 0)
			So(err.Error(), ShouldContainSubstring, "No file path set - no metrics are collected")
		})

		Convey("I should receive no metrics and error when SPECjbb results parsing fails", func() {
			configuration := makeDefaultConfiguration("specjbb_incorrect_format.stdout")
			metricTypes[0].Config = configuration

			metrics, err := specjbbPlugin.CollectMetrics(metricTypes)

			So(metrics, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
		})
	})
}

func makeDefaultConfiguration(fileName string) plugin.Config {
	configuration := plugin.Config{}
	configuration["stdout_file"] = fileName
	configuration["phase_name"] = "phase name"
	configuration["experiment_name"] = "experiment name"

	return configuration
}

func soValidMetricType(metricType plugin.Metric, namespace string, unit string) {
	So("/"+strings.Join(metricType.Namespace.Strings(), "/"), ShouldEqual, namespace)
	So(metricType.Unit, ShouldEqual, unit)
	So(metricType.Version, ShouldEqual, 1)
}

func soValidMetric(metric plugin.Metric, expectedMetric metric) {
	namespaceSuffix := expectedMetric.namespace
	value := expectedMetric.value
	time := expectedMetric.date

	namespace := "/" + strings.Join(metric.Namespace.Strings(), "/")
	So(namespace, ShouldStartWith, "/intel/swan/specjbb/")
	So(namespace, ShouldEndWith, namespaceSuffix)
	So(strings.Contains(namespace, "*"), ShouldBeFalse)
	So(metric.Unit, ShouldEqual, "ns")
	So(metric.Tags, ShouldHaveLength, 0)
	data, typeFound := metric.Data.(uint64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
	So(metric.Timestamp.Unix(), ShouldEqual, time.Unix())
}
