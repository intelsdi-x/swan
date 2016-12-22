package specjbb

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/parser"
	"github.com/pkg/errors"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME    = "specjbb"
	VERSION = 1
	TYPE    = snapPlugin.CollectorPluginType
	UNIT    = "ns"
)

var (
	namespace = []string{"intel", "swan", "specjbb"}
)

type plugin struct {
	now time.Time
}

// NewSpecjbb creates new specjbb collector.
func NewSpecjbb(now time.Time) snapPlugin.CollectorPlugin {
	return &plugin{now}
}

// GetMetricTypes implements plugin.PluginCollector interface.
func (specjbb *plugin) GetMetricTypes(configType snapPlugin.ConfigType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	metricNames := [][]string{
		[]string{"min"},
		[]string{"max"},
		[]string{"percentile", "50th"},
		[]string{"percentile", "90th"},
		[]string{"percentile", "95th"},
		[]string{"percentile", "99th"},
		[]string{"qps"},
		[]string{"issued_requests"}}

	for _, metricName := range metricNames {
		metrics = append(metrics, snapPlugin.MetricType{Namespace_: createNewMetricNamespace(metricName...), Unit_: UNIT, Version_: VERSION})
	}

	return metrics, nil
}

func createNewMetricNamespace(metricName ...string) core.Namespace {
	namespace := core.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	for _, value := range metricName {
		namespace = namespace.AddStaticElement(value)
	}

	return namespace
}

// CollectMetrics implements plugin.PluginCollector interface.
func (specjbb *plugin) CollectMetrics(metricTypes []snapPlugin.MetricType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	sourceFilePath, err := config.GetConfigItem(metricTypes[0], "stdout_file")
	if err != nil {
		msg := fmt.Sprintf("No file path set - no metrics are collected: %s", err.Error())
		log.Error(msg)
		return metrics, errors.Wrap(err, msg)
	}

	rawMetrics, err := parser.FileWithLatencies(sourceFilePath.(string))
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
		metric := snapPlugin.MetricType{Namespace_: metricType.Namespace_, Unit_: metricType.Unit_, Version_: metricType.Version_}
		metric.Namespace_[namespaceHostnameIndex].Value = hostname
		metric.Timestamp_ = specjbb.now

		// Strips prefix. For example: '/intel/swan/specjbb/<hostname>/avg' to '/avg'.
		metricNamespaceSuffix := metric.Namespace_[swanNamespacePrefix:]

		// Convert slice of namespace elements to string, so ['percentile', '95th'] becomes 'percentile/95th'
		metricName := metricNamespaceSuffix[0].Value
		for _, namespace := range metricNamespaceSuffix[1:] {
			metricName = strings.Join([]string{metricName, namespace.Value}, "/")
		}
		if value, ok := rawMetrics.Raw[metricName]; ok {
			metric.Data_ = value
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil

}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (specjbb *plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	policy := cpolicy.New()
	stdoutFile, err := cpolicy.NewStringRule("stdout_file", true)
	if err != nil {
		return policy, errors.Wrap(err, "cannot create new string rule")
	}
	policyNode := cpolicy.NewPolicyNode()
	policyNode.Add(stdoutFile)
	policy.Add(namespace, policyNode)

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

	return meta
}
