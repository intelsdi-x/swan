package swanMetricsCollector

import (
	"encoding/json"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/swan/pkg/metrics"
	"fmt"
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

var storedMetrics []metrics.SwanMetrics

func updateMetrics(rawMetrics []byte) error {
	var metricsToStore metrics.SwanMetrics
	if err := json.Unmarshal(rawMetrics, &metricsToStore); err != nil {
		return err
	}

	if i := detectMetricIndex(metricsToStore.Tags); i > -1 {
		storedMetrics[i] = metricsToStore
	} else {
		storedMetrics = append(storedMetrics, metricsToStore)
	}

	return nil
}

func addDynamicNamespace() []plugin.MetricType {
	var mts []plugin.MetricType
	for i, _ := range storedMetrics {
		mts = append(mts, registerNewExperiment(i)...)
	}

	return mts
}

func registerNewExperiment(i int) []plugin.MetricType {
	var mts []plugin.MetricType
	for _, key := range metrics.GetMetricKeys() {
		mts = append(mts, plugin.MetricType{Namespace_: core.NewNamespace(generateNamespace(i, key)...)})
	}
	return mts
}

func detectMetricIndex(tag metrics.Tags) int {
	for i, _ := range storedMetrics {
		if storedMetrics[i].Tags.Compare(tag) {
			return i
		}
	}
	return -1
}

// CollectMetrics is an implementation needed for the Collector interface
// which returns all available metrics on demand.
func (f *SessionCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType

	mts = append(mts, addDynamicNamespace()...)

	for i, _ := range mts {
		ind, nsKey := getKeyFromNs(mts[i].Namespace())
		if ind < 0 {
			continue
		}

		val := getValue(ind, nsKey)
		if val == nil {
			continue
		}
		mts[i].Data_ = val
		mts[i].Timestamp_ = time.Now()
		metrics = append(metrics, mts[i])
	}

	return metrics, nil
}

// GetMetricTypes is an implementation needed for the Collector interface.
// Namespace has got following form: intel/swan/<Experiment_UUID>/<Phase_ID>/<Repetition_ID>/
func (f *SessionCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}
	namespace := core.NewNamespace("intel", "swan")
	mts = append(mts, plugin.MetricType{Namespace_: namespace})
	return mts, nil
}

func generateNamespace(i int, key string) []string {
	return []string{"intel",
		"swan",
		storedMetrics[i].Tags.ExperimentID(),
		storedMetrics[i].Tags.PhaseID(),
		storedMetrics[i].Tags.RepetitionID(),
		key}
}

func getValue(i int, key string) interface{} {
	val := storedMetrics[i].Metrics.GetValue(key)
	if val == nil || val == 0 || val == "" {
		return nil
	}
	return val
}

func getKeyFromNs(ns core.Namespace) (int, string) {
	nsPath := ns.Strings()
	// 4 means: experimentId, phaseId, repetitionId, metricName
	if len(nsPath) < 4 {
		return -1, ""
	}
	tag := metrics.NewTags(nsPath[len(nsPath)-4], nsPath[len(nsPath)-3], nsPath[len(nsPath)-2])
	return detectMetricIndex(tag), nsPath[len(nsPath)-1]
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
