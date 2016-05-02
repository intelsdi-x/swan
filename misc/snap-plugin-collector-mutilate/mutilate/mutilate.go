package mutilate

import (
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

// Constant representing plugin name, version, type and unit of measurement used
const (
	NAME    = "mutilate"
	VERSION = 1
	TYPE    = plugin.CollectorPluginType
	UNIT    = "ns"
)

type mutilate struct {
	now time.Time
}

// NewMutilate creates new mutilate collector
func NewMutilate(now time.Time) *mutilate {
	return &mutilate{now}
}

// GetMetricTypes implements plugin.PluginCollector interface
func (mutilate *mutilate) GetMetricTypes(configType plugin.ConfigType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("avg"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("std"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("min"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "5th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "10th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "90th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "95th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "99th"), Unit_: UNIT, Version_: VERSION})
	metrics = append(metrics, plugin.MetricType{Namespace_: createNewMetricNamespace("percentile", "99_999th"), Unit_: UNIT, Version_: VERSION})

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

// CollectMetrics implements plugin.PluginCollector interface
func (mutilate *mutilate) CollectMetrics(metricTypes []plugin.MetricType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType
	// TODO - pass file path in configuration
	//fmt.Printf("\nmetricTypes: %v\n", metricTypes[0])
	sourceFilePath, _ := config.GetConfigItem(metricTypes[0], "stdout_file")
	phaseName, err0 := config.GetConfigItem(metricTypes[0], "phase_name")
	experimentName, err1 := config.GetConfigItem(metricTypes[0], "experiment_name")
	fmt.Printf("%v, %v", err0, err1)
	//fmt.Printf("\nsourceFilePath: %v, %v\n", sourceFilePath, confErr)
	rawMetrics, err := parseMutilateStdout(sourceFilePath.(string))
	fmt.Printf("\nrawMetrics: %v\nerror: %v\n", rawMetrics, err)
	hostname, _ := os.Hostname()
	for key, metricType := range metricTypes {
		metricType.Data_ = rawMetrics[key].value
		metricType.Namespace_[3].Value = hostname
		metricType.Timestamp_ = mutilate.now
		metricType.Tags_["phase"] = phaseName.(string)
		metricType.Tags_["experiment"] = experimentName.(string)
		metrics = append(metrics, metricType)
	}

	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface
func (mutilate *mutilate) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
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
func Meta() *plugin.PluginMeta {
	meta := plugin.NewPluginMeta(
		NAME,
		VERSION,
		TYPE,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1100*time.Millisecond),
	)
	meta.RPCType = plugin.JSONRPC

	return meta
}
