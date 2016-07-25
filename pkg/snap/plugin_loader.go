package snap

import (
	"os"
	"path"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/pkg/conf"
)

const (
	// MutilateCollector is name of snap plugin binary.
	MutilateCollector string = "snap-plugin-collector-mutilate"
	// KubesnapDockerCollector is name of snap plugin binary.
	KubesnapDockerCollector = "kubesnap-plugin-collector-docker"

	// CassandraPublisher is name of snap plugin binary.
	CassandraPublisher = "snap-plugin-publisher-cassandra"
	// SessionPublisher is name of snap plugin binary.
	SessionPublisher = "snap-plugin-publisher-session-test"
)

var (
	goPath = os.Getenv("GOPATH")

	defaultPluginsPath = path.Join(goPath, "bin")

	snapdAddress = conf.NewStringFlag("snapd_address", "Address to snapd in `http://%s:%s` format", "http://127.0.0.1:8181")
	pluginsPath  = conf.NewFileFlag("snap_plugins_path", "Path to Snap Plugins directory", defaultPluginsPath)
)

// DefaultPluginLoaderConfig returns default config for PluginLoader.
func DefaultPluginLoaderConfig() PluginLoaderConfig {
	return PluginLoaderConfig{
		SnapdAddress: snapdAddress.Value(),
		PluginsPath:  pluginsPath.Value(),
	}
}

// PluginLoaderConfig contains configuration for PluginLoader.
type PluginLoaderConfig struct {
	SnapdAddress string
	PluginsPath  string
}

// PluginLoader is used to simplify Snap plugin loading.
type PluginLoader struct {
	pluginsClient *Plugins
	config        PluginLoaderConfig
}

// NewDefaultPluginLoader returns PluginLoader with DefaultPluginLoaderConfig.
// Returns error when could not connect to Snap.
func NewDefaultPluginLoader() (*PluginLoader, error) {
	return NewPluginLoader(DefaultPluginLoaderConfig())
}

// NewPluginLoader constructs PluginLoader with given config.
// Returns error when could not connect to Snap.
func NewPluginLoader(config PluginLoaderConfig) (*PluginLoader, error) {
	snapClient, err := client.New(config.SnapdAddress, "v1", true)
	if err != nil {
		return nil, err
	}
	plugins := NewPlugins(snapClient)

	return &PluginLoader{
		pluginsClient: plugins,
		config:        config,
	}, nil
}

// LoadPlugin loads selected plugin from plugin path.
func (l PluginLoader) LoadPlugin(plugin string) error {
	pluginName, pluginType := GetPluginNameAndType(plugin)

	isPluginLoaded, err := l.pluginsClient.IsLoaded(pluginType, pluginName)
	if err != nil {
		return err
	}

	if isPluginLoaded {
		return nil
	}

	pluginPath := path.Join(l.config.PluginsPath, plugin)
	return l.pluginsClient.LoadPlugin(pluginPath)
}
