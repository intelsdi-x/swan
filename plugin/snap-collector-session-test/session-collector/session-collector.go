package sessionCollector

import (
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

type SessionCollector struct{}

const (
	Name    = "session-test"
	Version = 1
	Type    = plugin.CollectorPluginType
)

var _ plugin.CollectorPlugin = (*SessionCollector)(nil)

func (f *SessionCollector) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := []plugin.PluginMetricType{}

	// Just keep emitting 1's
	// TODO(niklas): Normalize current time over total run time in x and use sin(x) instead.
	for i, _ := range mts {
		mts[i].Data_ = 1
		mts[i].Source_, _ = os.Hostname()
		mts[i].Timestamp_ = time.Now()
		metrics = append(metrics, mts[i])
	}

	return metrics, nil
}

func (f *SessionCollector) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	mts := []plugin.PluginMetricType{}

	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "swan", "session", "metric1"}})
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "swan", "session", "metric2"}})

	return mts, nil
}

func (f *SessionCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		Name,
		Version,
		Type,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1100*time.Millisecond),
	)
}
