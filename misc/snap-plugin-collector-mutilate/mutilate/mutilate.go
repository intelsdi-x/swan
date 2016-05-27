package mutilate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/logger"
	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME    = "mutilate"
	VERSION = 1
	TYPE    = snapPlugin.CollectorPluginType
	UNIT    = "ns"
)

type plugin struct {
	now time.Time
}

// NewMutilate creates new mutilate collector.
func NewMutilate(now time.Time) snapPlugin.CollectorPlugin {
	return &plugin{now}
}

// GetMetricTypes implements plugin.PluginCollector interface.
func (mutilate *plugin) GetMetricTypes(configType snapPlugin.ConfigType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("avg"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("std"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("min"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "5th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "10th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "90th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "95th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "99th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("qps", "total"), Unit_: UNIT, Version_: VERSION})

	customNamespace := createNewMetricNamespace("percentile")
	customNamespace = customNamespace.AddDynamicElement("percentile", "Custom percentile from mutilate").AddStaticElement("custom")

	metrics = append(metrics, snapPlugin.MetricType{Namespace_: customNamespace, Unit_: UNIT, Version_: VERSION})

	return metrics, nil
}

func createNewMetricNamespace(metricName ...string) core.Namespace {
	namespace := core.NewNamespace("intel", "swan", "mutilate")
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	for _, value := range metricName {
		namespace = namespace.AddStaticElement(value)
	}

	return namespace
}

// CollectMetrics implements plugin.PluginCollector interface.
func (mutilate *plugin) CollectMetrics(metricTypes []snapPlugin.MetricType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	sourceFilePath, err := config.GetConfigItem(metricTypes[0], "stdout_file")
	if err != nil {
		msg := fmt.Sprintf("No file path set - no metrics are collected: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	rawMetrics, err := parse(sourceFilePath.(string))
	if err != nil {
		msg := fmt.Sprintf("Mutilate output parsing failed: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	hostname, err := os.Hostname()
	if err != nil {
		msg := fmt.Sprintf("Cannot determine hostname: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	// Swan provides a patched version of mutilate, which let's a user provide
	// a custom percentile value for mutilate to report. By default, Mutilate
	// reports p5, p10, p90, p95 and p99.
	customPercentile := 0.0
	for metric := range rawMetrics {
		var percentile float64
		if n, err := fmt.Sscanf(metric, "percentile/%fth/custom", &percentile); err == nil && n == 1 {
			customPercentile = percentile
			break
		}
	}

	const namespaceHostnameIndex = 3
	swanNamespacePrefix := []string{"intel", "swan", "mutilate", "hostname"}

	for _, metricType := range metricTypes {
		metric := snapPlugin.MetricType{Namespace_: metricType.Namespace_, Unit_: metricType.Unit_, Version_: metricType.Version_}
		metric.Namespace_[namespaceHostnameIndex].Value = hostname
		metric.Timestamp_ = mutilate.now

		// Strips prefix. For example: '/intel/swan/mutilate/<hostname>/avg' to '/avg'.
		metricPostfix := metric.Namespace_[len(swanNamespacePrefix):]

		// Convert slice of namespace elements to slice of strings.
		metricPostfixStrings := []string{}
		for _, namespace := range metricPostfix {
			metricPostfixStrings = append(metricPostfixStrings, namespace.Value)
		}

		// Flatten to string so ['percentile', '5th'] becomes '/percentile/5th'.
		metricName := strings.Join(metricPostfixStrings, "/")

		if value, ok := rawMetrics[metricName]; ok {
			metric.Data_ = value
		} else {
			// If metric wasn't found directly in raw metrics map, it may be a custom
			// percentile latency. For example, 'percentile/*/custom'.
			// If the rawMetrics gathered contains, for example 'percentile/99.999th',
			// we provide it here.
			if customPercentile > 0.0 &&
				len(metricPostfix) == 3 &&
				metricPostfix[0].Value == "percentile" &&
				metricPostfix[1].Name == "percentile" && metricPostfix[1].Value == "*" &&
				metricPostfix[2].Value == "custom" {

				customPercentileKey := fmt.Sprintf("percentile/%2.3fth/custom", customPercentile)
				if value, ok := rawMetrics[customPercentileKey]; ok {
					// Warning: This namespace transformation is kept temporarily.
					// It is not recommended to change the namespace of a metric which is made
					// available to the metrics catalog.
					percentileString := fmt.Sprintf("%2.3fth", customPercentile)

					// Snap namespaces may not have '.' so we have to replace with "_".
					percentileStringSanitized := strings.Replace(percentileString, ".", "_", 1)

					// Specialize '*' to for example '99_999th'.
					metricPostfix[1].Value = percentileStringSanitized

					metric.Data_ = value
				} else {
					logger.LogError("Could not find raw metric for key '%s': skipping metric", customPercentileKey)
				}
			}
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil

}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (mutilate *plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	stdoutFile, _ := cpolicy.NewStringRule("stdout_file", true)
	experiment := cpolicy.NewPolicyNode()
	experiment.Add(stdoutFile)
	policy := cpolicy.New()
	policy.Add([]string{""}, experiment)

	return policy, nil
}

// Meta returns plugin metadata.
func Meta() *snapPlugin.PluginMeta {
	meta := snapPlugin.NewPluginMeta(
		NAME,
		VERSION,
		TYPE,
		[]string{snapPlugin.SnapGOBContentType},
		[]string{snapPlugin.SnapGOBContentType},
		snapPlugin.Unsecure(true),
		snapPlugin.RoutingStrategy(snapPlugin.DefaultRouting),
		snapPlugin.CacheTTL(1*time.Second),
	)
	meta.RPCType = snapPlugin.JSONRPC

	return meta
}
