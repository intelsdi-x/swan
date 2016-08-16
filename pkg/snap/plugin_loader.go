package snap

import (
	"os"
	"path"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
)

const (
	// KubesnapDockerCollector is name of snap plugin binary.
	KubesnapDockerCollector = "kubesnap-plugin-collector-docker"
	// MutilateCollector is name of snap plugin binary.
	MutilateCollector string = "snap-plugin-collector-mutilate"
	// SessionCollector is name of snap plugin binary.
	SessionCollector string = "snap-plugin-collector-session-test"

	// TagProcessor is name of snap plugin binary.
	TagProcessor string = "snap-plugin-processor-tag"

	// CassandraPublisher is name of snap plugin binary.
	CassandraPublisher = "snap-plugin-publisher-cassandra"
	// SessionPublisher is name of snap plugin binary.
	SessionPublisher = "snap-plugin-publisher-session-test"
)

var (
	goPath = os.Getenv("GOPATH")

	defaultPluginsPath = path.Join(goPath, "bin")

	snapdAddress = conf.NewStringFlag("snapd_address", "Address to snapd in `http://%s:%s` format", "http://127.0.0.1:8181")
	pluginsPath  = conf.NewDirFlag("snap_plugins_path", "Path to Snap Plugins directory", defaultPluginsPath)
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

// Load loads supplied plugin names from plugin path and returns slice of
// encountered errors.
func (l PluginLoader) Load(plugins ...string) error {
	var errors errcollection.ErrorCollection
	for _, plugin := range plugins {
		err := l.load(plugin)
		errors.Add(err)
	}
	return errors.GetErrIfAny()
}

// load loads selected plugin from plugin path.
func (l PluginLoader) load(plugin string) error {
	pluginName, pluginType, err := GetPluginNameAndType(plugin)
	if err != nil {
		return err
	}

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
