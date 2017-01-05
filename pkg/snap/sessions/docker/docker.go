package docker

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

// DefaultConfig returns default configuration for Docker Session Launcher.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", snap.PluginAnyVersion)
	sessions.ApplyCassandraConfiguration(publisher)

	return Config{
		SnapteldAddress: snap.SnapteldHTTPEndpoint.Value(),
		Interval:        1 * time.Second,
		Publisher:       publisher,
	}
}

// Config contains configuration for Docker Session Launcher.
type Config struct {
	SnapteldAddress string
	Publisher       *wmap.PublishWorkflowMapNode
	Interval        time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering Kubernetes Docker containers metrics.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client
}

// NewSessionLauncher constructs Docker Session Launcher.
func NewSessionLauncher(config Config) (*SessionLauncher, error) {
	snapClient, err := client.New(config.SnapteldAddress, "v1", true)
	if err != nil {
		return nil, err
	}

	loaderConfig := snap.DefaultPluginLoaderConfig()
	loaderConfig.SnapteldAddress = config.SnapteldAddress
	loader, err := snap.NewPluginLoader(loaderConfig)
	if err != nil {
		return nil, err
	}

	err = loader.Load(snap.DockerCollector, snap.CassandraPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession([]string{"/intel/docker/*/stats/cgroups/*"},
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

	// Start session.
	err := s.session.Start(tags)
	if err != nil {
		return nil, err
	}

	return s.session, nil
}
