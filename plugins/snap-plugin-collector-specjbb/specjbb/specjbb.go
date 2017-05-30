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
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/parser"
	"github.com/pkg/errors"
)

// Constants representing collector name, version, type and unit of measurement used.
const (
	NAME    = "specjbb"
	VERSION = 1
	UNIT    = "ns"
)

var (
	namespace = []string{"intel", "swan", "specjbb"}
)

type collector struct {
	now time.Time
}

// NewSpecjbb creates new specjbb collector.
func NewSpecjbb(now time.Time) plugin.Collector {
	return &collector{now}
}

// GetMetricTypes implements collector.PluginCollector interface.
func (specjbb *collector) GetMetricTypes(configType plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	metricNames := [][]string{
		{"min"},
		{"max"},
		{"percentile", "50th"},
		{"percentile", "90th"},
		{"percentile", "95th"},
		{"percentile", "99th"},
		{"qps"},
		{"issued_requests"}}

	for _, metricName := range metricNames {
		metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace(metricName...), Unit: UNIT, Version: VERSION})
	}

	return metrics, nil
}

func createNewMetricNamespace(metricName ...string) plugin.Namespace {
	namespace := plugin.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	for _, value := range metricName {
		namespace = namespace.AddStaticElement(value)
	}

	return namespace
}

// CollectMetrics implements collector.PluginCollector interface.
func (specjbb *collector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	var metrics []plugin.Metric

	sourceFileName, err := metricTypes[0].Config.GetString("stdout_file")
	if err != nil {
		msg := fmt.Sprintf("No file path set - no metrics are collected: %s", err.Error())
		log.Error(msg)
		return metrics, errors.Wrap(err, msg)
	}

	rawMetrics, err := parser.FileWithLatencies(sourceFileName)
	if err != nil {
		msg := fmt.Sprintf("SPECjbb output parsing failed: %s", err.Error())
		log.Error(msg)
		return metrics, errors.Wrap(err, msg)
	}
	hostname, err := os.Hostname()
	if err != nil {
		msg := fmt.Sprintf("Cannot determine hostname: %s", err.Error())
		log.Error(msg)
		return metrics, errors.Wrap(err, msg)
	}

	// NamespacePrefix has 4 elements {"intel", "swan", "specjbb", "hostname"}.
	const namespaceHostnameIndex = 3
	const swanNamespacePrefix = 4

	for _, metricType := range metricTypes {
		metric := plugin.Metric{Namespace: metricType.Namespace, Unit: metricType.Unit, Version: metricType.Version}
		metric.Namespace[namespaceHostnameIndex].Value = hostname
		metric.Timestamp = specjbb.now

		// Strips prefix. For example: '/intel/swan/specjbb/<hostname>/avg' to '/avg'.
		metricNamespaceSuffix := metric.Namespace[swanNamespacePrefix:]

		// Convert slice of namespace elements to string, so ['percentile', '95th'] becomes 'percentile/95th'
		metricName := metricNamespaceSuffix[0].Value
		for _, namespace := range metricNamespaceSuffix[1:] {
			metricName = strings.Join([]string{metricName, namespace.Value}, "/")
		}
		if value, ok := rawMetrics.Raw[metricName]; ok {
			metric.Data = value
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil

}

// GetConfigPolicy implements collector.PluginCollector interface.
func (specjbb *collector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	err := policy.AddNewStringRule(namespace, "stdout_file", true)
	if err != nil {
		return *policy, errors.Wrap(err, "cannot create new string rule")
	}

	return *policy, nil
}
