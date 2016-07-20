package snap

import (
	"os"
	"path"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/pkg/errors"
)

// Plugin is an enum specifying plugin to load.
type Plugin int

const (
	// MutilateCollector is an enum specifying plugin to load.
	MutilateCollector Plugin = iota
	// KubesnapDockerCollector is an enum specifying plugin to load.
	KubesnapDockerCollector

	// CassandraPublisher is an enum specifying plugin to load.
	CassandraPublisher
)

// DefaultPluginLoaderConfig returns default config for PluginLoader.
func DefaultPluginLoaderConfig() PluginLoaderConfig {
	goPath := os.Getenv("GOPATH")

	defaultMutilateCollectorPath := path.Join(fs.GetSwanBuildPath(), "snap-plugin-collector-mutilate")
	defaultKubesnapCollectorPath := path.Join(goPath, "bin", "kubesnap-plugin-collector-docker")
	defaultCassandraPublisherPath := path.Join(path.Join(goPath, "bin", "snap-plugin-publisher-cassandra"))

	return PluginLoaderConfig{
		SnapdAddress:            conf.NewStringFlag("snapd_address", "Address to snapd in `http://%s:%s` format", "http://127.0.0.1:8181").Value(),
		MutilateCollectorPath:   conf.NewFileFlag("mutilate_collector_path", "Path to Mutilate collector binary", defaultMutilateCollectorPath).Value(),
		KubernetesCollectorPath: conf.NewFileFlag("kubesnap_docker_collector_path", "Path to Kubesnap Docker collector binary", defaultKubesnapCollectorPath).Value(),
		CassandraPublisherPath:  conf.NewFileFlag("cassandra_publisher_path", "Path to Cassandra publisher binary", defaultCassandraPublisherPath).Value(),
	}
}

// PluginLoaderConfig contains configuration for PluginLoader.
type PluginLoaderConfig struct {
	SnapdAddress string

	MutilateCollectorPath   string
	KubernetesCollectorPath string

	CassandraPublisherPath string
}

// PluginLoader is used to simplify Snap plugin loading.
type PluginLoader struct {
	plugins *Plugins
	config  PluginLoaderConfig
}

// NewDefaultPluginLoader returns PluginLoader with DefaultPluginLoaderConfig.
// Returns error when could not connect to Snap.
func NewDefaultPluginLoader() (*PluginLoader, error) {
	return NewPluginLoader(DefaultPluginLoaderConfig())
}

// NewPluginFactory constructs PluginLoader with given config.
// Returns error when could not connect to Snap.
func NewPluginLoader(config PluginLoaderConfig) (*PluginLoader, error) {
	snapClient, err := client.New(config.SnapdAddress, "v1", true)
	if err != nil {
		return nil, err
	}
	plugins := NewPlugins(snapClient)

	return &PluginLoader{
		plugins: plugins,
		config:  config,
	}, nil
}

// LoadPlugin loads selected plugin.
func (f PluginLoader) LoadPlugin(plugin Plugin) error {
	switch plugin {
	case MutilateCollector:
		return f.plugins.LoadPlugin(f.config.MutilateCollectorPath)
	case KubesnapDockerCollector:
		return f.plugins.LoadPlugin(f.config.KubernetesCollectorPath)
	case CassandraPublisher:
		return f.plugins.LoadPlugin(f.config.CassandraPublisherPath)

	default:
		return errors.Errorf("plugin %q is not available", plugin)
	}
}
