package mutilate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/parse"
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

	customNamespace := createNewMetricNamespace("percentile")
	customNamespace = customNamespace.AddDynamicElement("percentile", "Custom percentile from mutilate").AddStaticElement("custom")

	metrics = append(metrics, plugin.Metric{Namespace: customNamespace, Unit: UNIT, Version: VERSION})

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
			// If metric wasn't found directly in raw metrics map, it may be a custom
			// percentile latency. For example, 'percentile/*/custom'.
			// If the rawMetrics gathered contains non-default LatencyPercentile field e.g '99.99900',
			// we provide it here.
			// Swan provides a patched version of mutilate, which let's a user provide
			// a custom percentile value for mutilate to report. By default, Mutilate
			// reports p5, p10, p90, p95 and p99.
			if rawMetrics.LatencyPercentile != "" &&
				len(metricNamespaceSuffix) == 3 &&
				metricNamespaceSuffix[0].Value == "percentile" &&
				metricNamespaceSuffix[1].Value == "*" && metricNamespaceSuffix[1].Name == "percentile" &&
				metricNamespaceSuffix[2].Value == "custom" {

				if value, ok := rawMetrics.Raw[parse.MutilatePercentileCustom]; ok {
					// Snap namespaces may not have '.' so we have to replace with "_".
					percentileStringSanitized := strings.Replace(
						fmt.Sprintf("%sth", rawMetrics.LatencyPercentile), ".", "_", 1)

					// Specialize '*' to for example '99_999th'.
					metricNamespaceSuffix[1].Value = percentileStringSanitized

					metric.Data = value
				} else {
					log.Errorf("Could not find raw metric for key '%s': skipping metric", parse.MutilatePercentileCustom)
				}
			} else {
				log.Errorf("Could not find raw metric for key '%s': skipping metric", metricName)
			}
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

// Meta returns collector metadata.
//func Meta() *snapPlugin.PluginMeta {
//	meta := snapPlugin.NewPluginMeta(
//		NAME,
//		VERSION,
//		TYPE,
//		[]string{snapPlugin.SnapGOBContentType},
//		[]string{snapPlugin.SnapGOBContentType},
//		snapPlugin.Unsecure(true),
//		snapPlugin.RoutingStrategy(snapPlugin.DefaultRouting),
//		snapPlugin.CacheTTL(1*time.Second),
//	)
//
//	return meta
//}
