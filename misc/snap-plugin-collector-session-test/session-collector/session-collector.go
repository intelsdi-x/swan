package sessionCollector

import (
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

// SessionCollector is a plugin which provides a Swan hosted mock colllector
// which simply emits metric value '1' in the /intel/swan/session/metric1 namespace.
type SessionCollector struct{}

const (
	name = "session-test"
	version = 1
	pluginType = plugin.CollectorPluginType
)

var _ plugin.CollectorPlugin = (*SessionCollector)(nil)

// CollectMetrics is an implementation needed for the Collector interface and here,
// simply just returns '1' for all metric types.
func (f *SessionCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	metrics := []plugin.MetricType{}

	// Just keep emitting 1's
	for i := range mts {
		mts[i].Data_ = 1
		mts[i].Timestamp_ = time.Now()
		metrics = append(metrics, mts[i])
	}

	return metrics, nil
}

// GetMetricTypes is an implementation needed for the Collector interface and here,
// simply just returns a static namespace /intel/swan/session/metric1.
func (f *SessionCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}

	mts = append(mts, plugin.MetricType{Namespace_: core.NewNamespace("intel", "swan", "session", "metric1")})

	return mts, nil
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
