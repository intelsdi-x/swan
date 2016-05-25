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
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("qps", "peak"), Unit_: UNIT, Version_: VERSION})

	customNamespace := createNewMetricNamespace("percentile")
	customNamespace = customNamespace.AddDynamicElement("percentile", "Custom percentile from mutilate").AddStaticElement("custom")

	metrics = append(metrics, snapPlugin.MetricType{Namespace_: customNamespace, Unit_: UNIT, Version_: VERSION})

	return metrics, nil
}

func createNewMetricNamespace(metricName ...string) core.Namespace {
	namespace := core.NewNamespace("intel", "swan", "mutilate")
	namespace = namespace.AddDynamicElement("hostname",
		"Name of the host that reports the metric")
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

	rawMetrics, err := parseMutilateStdout(sourceFilePath.(string))
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

	// TODO: Refactor population below. It is hard to figure out what is going on.
	for key, metricType := range metricTypes {
		logger.LogError("key: %s", key)
		logger.LogError("metricType: %v", metricType)

		metric := snapPlugin.MetricType{Namespace_: metricType.Namespace_, Unit_: metricType.Unit_, Version_: metricType.Version_}
		metric.Namespace_[3].Value = hostname
		metric.Timestamp_ = mutilate.now

		for _, m := range rawMetrics {
			// Assign value to metric for proper name.
			if strings.Contains(metricType.Namespace().String(), m.name) {
				metric.Data_ = m.value
			}
		}

		// Namespace: ['percentile', *, 'custom']
		dynamicNamespace := metric.Namespace_[len(metric.Namespace_)-2]
		logger.LogError("dynamicNamespace: %v", dynamicNamespace)
		logger.LogError("rawMetrics: %v", rawMetrics)
		if dynamicNamespace.Name == "percentile" && dynamicNamespace.Value == "*" {
			metric.Namespace_[len(metric.Namespace_)-2].Value = strings.Replace(strings.Split(rawMetrics[7].name, "/")[1], ".", "_", -1)

			// Assign value for custom metric - always last item in rawMetrics.
			metric.Data_ = rawMetrics[len(rawMetrics)-1].value
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
