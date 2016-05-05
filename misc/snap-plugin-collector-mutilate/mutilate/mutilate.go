package mutilate

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/logger"
	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

// Constant representing plugin name, version, type and unit of measurement used
const (
	NAME    = "mutilate"
	VERSION = 1
	TYPE    = snapPlugin.CollectorPluginType
	UNIT    = "ns"
)

type plugin struct {
	now time.Time
}

// NewMutilate creates new mutilate collector
func NewMutilate(now time.Time) snapPlugin.CollectorPlugin {
	return &plugin{now}
}

// GetMetricTypes implements plugin.PluginCollector interface
func (mutilate *plugin) GetMetricTypes(configType snapPlugin.ConfigType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("avg"),
		Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("std"),
		Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace("min"),
		Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "5th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "10th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "90th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "95th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "99th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(
		"percentile", "99_999th"), Unit_: UNIT, Version_: VERSION})

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

// CollectMetrics implements plugin.PluginCollector interface
func (mutilate *plugin) CollectMetrics(metricTypes []snapPlugin.MetricType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType
	sourceFilePath, sFPErr := config.GetConfigItem(metricTypes[0], "stdout_file")
	if sFPErr != nil {
		logger.LogError("No file path set - no metrics are collected", sFPErr)
		return metrics, errors.New("No file path set - no metrics are collected")
	}
	phaseName, pNErr := config.GetConfigItem(metricTypes[0], "phase_name")
	if pNErr != nil {
		logger.LogError("No phase name set - no metrics are collected", pNErr)
		return metrics, errors.New("No phase name set - no metrics are collected")
	}
	experimentName, eNError := config.GetConfigItem(metricTypes[0], "experiment_name")
	if eNError != nil {
		logger.LogError("No experiment name set - no metrics are collected", eNError)
		return metrics, errors.New("No experiment name set - no metrics are collected")
	}
	rawMetrics, rMErr := parseMutilateStdout(sourceFilePath.(string))
	if rMErr != nil {
		logger.LogError(fmt.Sprintf("Mutilate output parsing failed: %s", rMErr.Error()), rMErr)
		return metrics, fmt.Errorf("Mutilate output parsing failed: %s", rMErr.Error())
	}
	hostname, hErr := os.Hostname()
	if hErr != nil {
		logger.LogError("Can not determine hostname", hErr)
		return metrics, fmt.Errorf("Can not determine hostname: %s", hErr.Error())
	}

	for key, metricType := range metricTypes {
		metric := snapPlugin.MetricType{Namespace_: metricType.Namespace_,
			Unit_: metricType.Unit_, Version_: metricType.Version_}
		metric.Data_ = rawMetrics[key].value
		metric.Namespace_[3].Value = hostname
		metric.Timestamp_ = mutilate.now
		metric.Tags_ = map[string]string{"phase": phaseName.(string),
			"experiment": experimentName.(string)}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface
func (mutilate *plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	phaseName, _ := cpolicy.NewStringRule("phase_name", true)
	stdoutFile, _ := cpolicy.NewStringRule("stdout_file", true)
	experimentName, _ := cpolicy.NewStringRule("experiment_name", true)
	experiment := cpolicy.NewPolicyNode()
	experiment.Add(phaseName)
	experiment.Add(stdoutFile)
	experiment.Add(experimentName)
	policy := cpolicy.New()
	policy.Add([]string{""}, experiment)

	return policy, nil
}

// Meta returns plugin metadata
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
