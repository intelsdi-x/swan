package kubesnap

import (
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
)

// DefaultConfig returns default configuration for Kubesnap Session Launcher.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", 2)
	publisher.AddConfigItem("server", conf.CassandraAddress.Value())

	return Config{
		SnapdAddress: snap.SnapdHTTPEndpoint.Value(),
		Interval:     1 * time.Second,
		Publisher:    publisher,
	}
}

// Config contains configuration for Kubesnap Session Launcher.
type Config struct {
	SnapdAddress string
	Publisher    *wmap.PublishWorkflowMapNode
	Interval     time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering Kubernetes Docker containers metrics.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client
}

// NewSessionLauncher constructs Kubesnap Session Launcher.
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

	err = loader.Load(snap.KubesnapDockerCollector, snap.CassandraPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession([]string{
			// "/intel/docker/*/cgroups/*"
			"/intel/docker/*", // should include labels
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
	phaseSession phase.Session) (snap.SessionHandle, error) {

	// Start session.
	err := s.session.Start(phaseSession)
	if err != nil {
		return nil, err
	}

	return s.session, nil
}
