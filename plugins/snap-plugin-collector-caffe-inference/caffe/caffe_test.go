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

package caffe

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
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

		Convey("I should receive information about metrics", func() {
			metricTypes, err := caffePlugin.GetMetricTypes(plugin.Config{})
			So(err, ShouldBeNil)
			So(metricTypes, ShouldHaveLength, expectedMetricsCount)
			soValidMetricType(metricTypes[0], "/intel/swan/caffe/inference/*/batches", METRICNAME)

			Convey("I should receive valid metrics when I try to collect them", func() {
				configuration := makeDefaultConfiguration("log-finished.txt")
				metricTypes[0].Config = configuration
				collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(err, ShouldBeNil)
				So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
				So(strings.Join(collectedMetrics[0].Namespace.Strings(), "/"), ShouldContainSubstring, expectedMetric.namespace)
				expectedMetric.value = 99
				soValidMetric(collectedMetrics[0], expectedMetric)

			})
			Convey("I should receive no metrics end error when caffe ended prematurely ", func() {
				configuration := makeDefaultConfiguration("log-notstarted.txt")
				metricTypes[0].Config = configuration
				collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(collectedMetrics, ShouldHaveLength, 0)
				So(err, ShouldEqual, ErrParse)
			})
			Convey("I should receive no metrics end error when caffe ended prematurely and there is single work 'Batch' in log without number", func() {
				configuration := makeDefaultConfiguration("log-interrupted2.txt")
				metricTypes[0].Config = configuration
				collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(collectedMetrics, ShouldHaveLength, 0)
				So(err, ShouldEqual, ErrParse)
			})
			Convey("I should receive valid metric when caffe was killed during inference", func() {
				configuration := makeDefaultConfiguration("log-interrupted.txt")
				metricTypes[0].Config = configuration
				collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(err, ShouldBeNil)
				So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
				So(strings.Join(collectedMetrics[0].Namespace.Strings(), "/"), ShouldContainSubstring, expectedMetric.namespace)
				expectedMetric.value = 24
				soValidMetric(collectedMetrics[0], expectedMetric)
			})
			Convey("I should receive valid metric when caffe was killed during inference and last word in log is 'Batch'", func() {
				configuration := makeDefaultConfiguration("log-interrupted3.txt")
				metricTypes[0].Config = configuration
				collectedMetrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(err, ShouldBeNil)
				So(collectedMetrics, ShouldHaveLength, expectedMetricsCount)
				So(strings.Join(collectedMetrics[0].Namespace.Strings(), "/"), ShouldContainSubstring, expectedMetric.namespace)
				expectedMetric.value = 3
				soValidMetric(collectedMetrics[0], expectedMetric)
			})
			Convey("I should receive no metrics and error when no file path is set", func() {
				configuration := plugin.Config{}
				metricTypes[0].Config = configuration
				metrics, err := caffePlugin.CollectMetrics(metricTypes)
				So(metrics, ShouldHaveLength, 0)
				So(err, ShouldEqual, ErrConf)
			})
		})


	})
}

func makeDefaultConfiguration(fileName string) plugin.Config {
	configuration := plugin.Config{}
	configuration["stdout_file"] = fileName
	return configuration
}

func soValidMetricType(metricType plugin.Metric, namespace string, unit string) {
	So(strings.Join(append([]string{""}, metricType.Namespace.Strings()...), "/"), ShouldEqual, namespace)
	So(metricType.Unit, ShouldEqual, unit)
	So(metricType.Version, ShouldEqual, 1)
}

func soValidMetric(metric plugin.Metric, expectedMetric metric) {
	namespaceSuffix := expectedMetric.namespace
	namespacePrefix := strings.Join(append([]string{""}, namespace...), "/")
	value := expectedMetric.value

	namespace := strings.Join(append([]string{""}, metric.Namespace.Strings()...), "/")
	So(namespace, ShouldStartWith, namespacePrefix)
	So(namespace, ShouldEndWith, namespaceSuffix)
	So(strings.Contains(namespace, "*"), ShouldBeFalse)
	So(metric.Unit, ShouldEqual, METRICNAME)
	So(metric.Tags, ShouldHaveLength, 0)
	data, typeFound := metric.Data.(uint64)
	So(typeFound, ShouldBeTrue)
	So(data, ShouldEqual, value)
}
