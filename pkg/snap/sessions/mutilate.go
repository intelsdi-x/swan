package sessions

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	"path"
	"time"
)

// MutilateSnapSessionLauncher configures & launches snap workflow for gathering
// SLIs from Mutilate.
type MutilateSnapSessionLauncher struct {
	session                    *snap.Session
	snapClient                 *client.Client
	mutilateCollectorBuildPath string
}

// NewMutilateSnapSessionLauncher constructs MutilateSnapSessionLauncher.
func NewMutilateSnapSessionLauncher(
	mutilateCollectorBuildPath string,
	interval time.Duration,
	snapClient *client.Client,
	publisher *wmap.PublishWorkflowMapNode) *MutilateSnapSessionLauncher {

	return &MutilateSnapSessionLauncher{
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
				//TODO: Fetch the 99_999th value from MUTILATE task itself!
				//Does not work for now:
				// "/intel/swan/mutilate/*/percentile/99_999th/custom",
			},
			interval,
			snapClient,
			publisher,
		),
		snapClient:                 snapClient,
		mutilateCollectorBuildPath: mutilateCollectorBuildPath,
	}
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *MutilateSnapSessionLauncher) LaunchSession(
	task executor.TaskInfo, phaseSession phase.Session) (snap.SessionHandle, error) {

	// Check if Mutilate collector plugin is loaded.
	plugins := snap.NewPlugins(s.snapClient)
	loaded, err := plugins.IsLoaded("collector", "mutilate")
	if err != nil {
		return nil, err
	}

	if !loaded {
		pluginPath :=
			[]string{path.Join(s.mutilateCollectorBuildPath, "snap-plugin-collector-mutilate")}
		err := plugins.Load(pluginPath)
		if err != nil {
			return nil, err
		}
	}

	// Getting FileName.
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
