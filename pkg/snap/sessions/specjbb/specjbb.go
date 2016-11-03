package specjbbsession

import (
	"time"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

// DefaultConfig returns default configuration for SPECjbb Collector session.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", 2)
	publisher.AddConfigItem("server", conf.CassandraAddress.Value())

	return Config{
		SnapdAddress: snap.SnapdHTTPEndpoint.Value(),
		Interval:     1 * time.Second,
		Publisher:    publisher,
	}
}

// Config contains configuration for SPECjbb Collector session.
type Config struct {
	SnapdAddress string
	Publisher    *wmap.PublishWorkflowMapNode
	Interval     time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering
// metrics from SPECjbb.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client
}

// NewSessionLauncher constructs SPECjbbSnapSessionLauncher.
func NewSessionLauncher(config Config) (*SessionLauncher, error) {
	snapClient, err := client.New(config.SnapdAddress, "v1", true)
	if err != nil {
		return nil, err
	}

	loaderConfig := snap.DefaultPluginLoaderConfig()
	loaderConfig.SnapdAddress = config.SnapdAddress
	loader, err := snap.NewPluginLoader(loaderConfig)
	if err != nil {
		return nil, err
	}

	err = loader.Load(snap.SPECjbbCollector, snap.CassandraPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession(
			[]string{
				"/intel/swan/specjbb/*/min",
				"/intel/swan/specjbb/*/percentile/50th",
				"/intel/swan/specjbb/*/percentile/90th",
				"/intel/swan/specjbb/*/percentile/95th",
				"/intel/swan/specjbb/*/percentile/99th",
				"/intel/swan/specjbb/*/max",
			},
			config.Interval,
			snapClient,
			config.Publisher,
		),
		snapClient: snapClient,
	}, nil
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *SessionLauncher) LaunchSession(
	task executor.TaskInfo,
	tags string) (snap.SessionHandle, error) {

	// Obtain SPECjbb output file.
	stdoutFile, err := task.StdoutFile()
	if err != nil {
		return nil, err
	}

	// Configuring SPECjbb collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		snap.CollectNodeConfigItem{
			Ns:    "/intel/swan/specjbb",
			Key:   "stdout_file",
			Value: stdoutFile.Name(),
		},
	}

	// Start session.
	err = s.session.Start(tags)
	if err != nil {
		return nil, err
	}

	return s.session, nil
}
