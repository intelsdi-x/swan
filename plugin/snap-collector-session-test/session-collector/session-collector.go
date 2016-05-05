package sessionCollector

import (
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"

	log "github.com/Sirupsen/logrus"
)

type SessionCollector struct{}

const (
	Name    = "session-test"
	Version = 1
	Type    = plugin.CollectorPluginType
)

var _ plugin.CollectorPlugin = (*SessionCollector)(nil)

func (f *SessionCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	logger := log.New()
	metrics := []plugin.MetricType{}

	// Just keep emitting 1's
	for i, _ := range mts {
		mts[i].Data_ = 1
		mts[i].Timestamp_ = time.Now()
		metrics = append(metrics, mts[i])

		logger.Printf("Emitted 1 at %s", mts[i].Timestamp_.String())
	}

	return metrics, nil
}

func (f *SessionCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}

	mts = append(mts, plugin.MetricType{Namespace_: core.NewNamespace("intel", "swan", "session", "metric1")})
	mts = append(mts, plugin.MetricType{Namespace_: core.NewNamespace("intel", "swan", "session", "metric2")})

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
