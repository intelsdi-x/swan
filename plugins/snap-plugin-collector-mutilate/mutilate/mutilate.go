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

package mutilate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-collector-mutilate/mutilate/parse"
)

// Constants representing collector name, version, type and unit of measurement used.
const (
	NAME    = "mutilate"
	VERSION = 1
	UNIT    = "ns"
)

type collector struct {
	now time.Time
}

// NewMutilate creates new mutilate collector.
func NewMutilate(now time.Time) plugin.Collector {
	return plugin.Collector(collector{now})
}

// GetMetricTypes implements collector.PluginCollector interface.
func (mutilate collector) GetMetricTypes(configType plugin.Config) ([]plugin.Metric, error) {
	var metrics []plugin.Metric

	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("avg"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("std"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("min"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("percentile", "5th"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("percentile", "10th"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("percentile", "90th"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("percentile", "95th"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("percentile", "99th"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("qps"), Unit: UNIT, Version: VERSION})
	metrics = append(metrics, plugin.Metric{Namespace: createNewMetricNamespace("misses"), Unit: UNIT, Version: VERSION})

	return metrics, nil
}

func createNewMetricNamespace(metricName ...string) plugin.Namespace {
	namespace := plugin.NewNamespace("intel", "swan", "mutilate")
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	for _, value := range metricName {
		namespace = namespace.AddStaticElement(value)
	}

	return namespace
}

// CollectMetrics implements collector.PluginCollector interface.
func (mutilate collector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	var metrics []plugin.Metric

	sourceFileName, err := metricTypes[0].Config.GetString("stdout_file")
	if err != nil {
		msg := fmt.Sprintf("No file path set - no metrics are collected: %s", err.Error())
		log.Error(msg)
		return metrics, errors.New(msg)
	}

	rawMetrics, err := parse.File(sourceFileName)
	if err != nil {
		msg := fmt.Sprintf("Mutilate output parsing failed: %s", err.Error())
		log.Error(msg)
		return metrics, errors.New(msg)
	}

	hostname, err := os.Hostname()
	if err != nil {
		msg := fmt.Sprintf("Cannot determine hostname: %s", err.Error())
		log.Error(msg)
		return metrics, errors.New(msg)
	}

	const namespaceHostnameIndex = 3
	const swanNamespacePrefix = len([...]string{"intel", "swan", "mutilate", "hostname"})

	for _, metricType := range metricTypes {
		metric := plugin.Metric{Namespace: metricType.Namespace, Unit: metricType.Unit, Version: metricType.Version}
		metric.Namespace[namespaceHostnameIndex].Value = hostname
		metric.Timestamp = mutilate.now

		// Strips prefix. For example: '/intel/swan/mutilate/<hostname>/avg' to '/avg'.
		metricNamespaceSuffix := metric.Namespace[swanNamespacePrefix:]

		// Convert slice of namespace elements to slice of strings.
		metricNamespaceSuffixStrings := []string{}
		for _, namespace := range metricNamespaceSuffix {
			metricNamespaceSuffixStrings = append(metricNamespaceSuffixStrings, namespace.Value)
		}

		// Flatten to string so ['percentile', '5th'] becomes '/percentile/5th'.
		metricName := strings.Join(metricNamespaceSuffixStrings, "/")

		if value, ok := rawMetrics.Raw[metricName]; ok {
			metric.Data = value
		} else {
			log.Errorf("Could not find raw metric for key '%s': skipping metric", metricName)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil

}

// GetConfigPolicy implements collector.PluginCollector interface.
func (mutilate collector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	err := policy.AddNewStringRule([]string{"intel", "swan", "mutilate"}, "stdout_file", true)
	if err != nil {
		return plugin.ConfigPolicy{}, err
	}

	return *policy, nil
}
