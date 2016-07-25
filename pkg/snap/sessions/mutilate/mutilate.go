package mutilatesession

import (
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
)

// DefaultConfig returns default configuration for Mutilate Collector session.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", 2)
	publisher.AddConfigItem("server", cassandra.AddrFlag.Value())

	return Config{
		SnapdAddress: snap.SnapdHTTPEndpoint.Value(),
		Interval:     1 * time.Second,
		Publisher:    publisher,
	}
}

// Config contains configuration for Mutilate Collector session.
type Config struct {
	SnapdAddress string
	Publisher    *wmap.PublishWorkflowMapNode
	Interval     time.Duration
}

// MutilateSnapSessionLauncher configures & launches snap workflow for gathering
// SLIs from Mutilate.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client
}

// NewMutilateSnapSessionLauncher constructs MutilateSnapSessionLauncher.
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

	err = loadPlugins(loader)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession(
			[]string{
				"/intel/swan/mutilate/*/avg",
				"/intel/swan/mutilate/*/std",
				"/intel/swan/mutilate/*/min",
				"/intel/swan/mutilate/*/percentile/5th",
				"/intel/swan/mutilate/*/percentile/10th",
				"/intel/swan/mutilate/*/percentile/90th",
				"/intel/swan/mutilate/*/percentile/95th",
				"/intel/swan/mutilate/*/percentile/99th",
				"/intel/swan/mutilate/*/qps",
				//TODO: Fetch the 99_999th value from MUTILATE task itself!
				//It shall be redesigned ASAP
				"/intel/swan/mutilate/*/percentile/*/custom",
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

	// Obtain Mutilate output file.
	stdoutFile, err := task.StdoutFile()
	if err != nil {
		return nil, err
	}

	// Configuring Mutilate collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		snap.CollectNodeConfigItem{
			Ns:    "/intel/swan/mutilate",
			Key:   "stdout_file",
			Value: stdoutFile.Name(),
		},
	}

	// Start session.
	err = s.session.Start(phaseSession)
	if err != nil {
		return nil, err
	}

	return s.session, nil
}

func loadPlugins(loader *snap.PluginLoader) (err error) {
	err = loader.LoadPlugin(snap.MutilateCollector)
	if err != nil {
		return err
	}

	err = loader.LoadPlugin(snap.CassandraPublisher)
	if err != nil {
		return err
	}

	return nil
}
