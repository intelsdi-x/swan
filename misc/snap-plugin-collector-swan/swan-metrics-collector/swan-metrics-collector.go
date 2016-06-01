package swanMetricsCollector

import (
	"encoding/json"
	"time"

	"fmt"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/swan/pkg/metrics"
)

// SessionCollector is a plugin which provides a Swan hosted mock collector
// which simply emits metric value '1' in the /intel/swan/session/metric1 namespace.
type SessionCollector struct{}

const (
	name       = "swan-metrics-collector"
	version    = 1
	pluginType = plugin.CollectorPluginType
)

var _ plugin.CollectorPlugin = (*SessionCollector)(nil)

type storedMetric struct {
	metric     metrics.Swan
	updateDate time.Time
}

var updateChannel = make(chan storedMetric)

func updateMetrics(rawMetrics []byte) error {
	var metricsToStore metrics.Swan

	if err := json.Unmarshal(rawMetrics, &metricsToStore); err != nil {
		return err
	}

	metricObject := storedMetric{metric: metricsToStore, updateDate: time.Now()}
	updateChannel <- metricObject

	return nil
}

type swanMetadataNamespace struct {
	namespace []string
	update    time.Time
}

func unpackStoredMetrics(swanMetrics storedMetric) (metrics []plugin.MetricType) {

	namespace := swanMetadataNamespace{
		namespace: generateNamespace(swanMetrics.metric.Tags),
		update:    swanMetrics.updateDate,
	}

	metrics = append(metrics, namespace.fulfillMetricObject("AggressorName", swanMetrics.metric.Metrics.AggressorName))
	metrics = append(metrics, namespace.fulfillMetricObject("AggressorParameters", swanMetrics.metric.Metrics.AggressorParameters))
	metrics = append(metrics, namespace.fulfillMetricObject("LCIsolation", swanMetrics.metric.Metrics.LCIsolation))
	metrics = append(metrics, namespace.fulfillMetricObject("LCName", swanMetrics.metric.Metrics.LCName))
	metrics = append(metrics, namespace.fulfillMetricObject("LCParameters", swanMetrics.metric.Metrics.LCParameters))
	metrics = append(metrics, namespace.fulfillMetricObject("LGIsolation", swanMetrics.metric.Metrics.LGIsolation))
	metrics = append(metrics, namespace.fulfillMetricObject("LGName", swanMetrics.metric.Metrics.LGName))
	metrics = append(metrics, namespace.fulfillMetricObject("LCParameters", swanMetrics.metric.Metrics.LCParameters))
	metrics = append(metrics, namespace.fulfillMetricObject("LoadDuration", swanMetrics.metric.Metrics.LoadDuration))
	metrics = append(metrics, namespace.fulfillMetricObject("TuningDuration", swanMetrics.metric.Metrics.TuningDuration))
	metrics = append(metrics, namespace.fulfillMetricObject("LoadPointsNumber", swanMetrics.metric.Metrics.LoadPointsNumber))
	metrics = append(metrics, namespace.fulfillMetricObject("RepetitionsNumber", swanMetrics.metric.Metrics.RepetitionsNumber))
	metrics = append(metrics, namespace.fulfillMetricObject("QPS", swanMetrics.metric.Metrics.QPS))
	metrics = append(metrics, namespace.fulfillMetricObject("SLO", swanMetrics.metric.Metrics.SLO))
	metrics = append(metrics, namespace.fulfillMetricObject("TargetLoad", swanMetrics.metric.Metrics.TargetLoad))

	return metrics
}

func (n swanMetadataNamespace) fulfillMetricObject(metricName string, value interface{}) plugin.MetricType {
	return plugin.MetricType{
		Namespace_: core.NewNamespace(append(n.namespace, metricName)...),
		Data_:      value,
		Timestamp_: n.update,
	}
}

// CollectMetrics is an implementation needed for the Collector interface
// which returns all available metrics on demand.
func (f *SessionCollector) CollectMetrics(mts []plugin.MetricType) (metrics []plugin.MetricType, err error) {

	select {
	case data := <-updateChannel:
		return append(metrics, unpackStoredMetrics(data)...), err
	default:
		return metrics, err
	}
}

// GetMetricTypes is an implementation needed for the Collector interface.
// Namespace has got following form: intel/swan/<Experiment_UUID>/<Phase_ID>/<Repetition_ID>/
func (f *SessionCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}
	namespace := core.NewNamespace("intel", "swan")
	mts = append(mts, plugin.MetricType{Namespace_: namespace})
	return mts, nil
}

func generateNamespace(swanTag metrics.Tags) (namespace []string) {
	namespace = append(namespace, []string{"intel", "swan"}...)
	if swanTag.ExperimentID != "" {
		namespace = append(namespace, swanTag.ExperimentID)
	}
	if swanTag.PhaseID != "" {
		namespace = append(namespace, swanTag.PhaseID)
	}
	if swanTag.RepetitionID > 0 {
		namespace = append(namespace, fmt.Sprintf("%d", swanTag.RepetitionID))
	}
	return namespace
}

// GetConfigPolicy is an implementation needed for the Collector interface and here,
// returns an empty configuration policy.
func (f *SessionCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

// Meta returns a plugin meta data.
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		name,
		version,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1100*time.Millisecond),
	)
}
