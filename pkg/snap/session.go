package snap

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"os"
	"path"
	"time"
)

type task struct {
	Version  int
	Schedule *client.Schedule
	Workflow *wmap.WorkflowMap
	Name     string
	Deadline string
	ID       string
	State    string
}

// Session provides construct for tagging metrics for a specified time span
// defined by Start() and Stop().
type Session struct {
	//	// Experiment is the **unique** experiment id. For example 'foobar-13a1b0bb-4467-4476-9818-986effe5c963'.
	//	Experiment string
	//
	//	// Phase is the **unique** phase id. For example 'barbaz-13a1b0bb-4467-4476-9818-986effe5c963'.
	//	Phase string

	// Interval defines the sample interval for the listed metrics.
	Interval time.Duration

	// Metrics to tag in session.
	Metrics []string

	// Active task.
	task *task

	// Publisher for tagged metrics.
	Publisher *wmap.PublishWorkflowMapNode

	// Client to Snapd.
	pClient *client.Client
}

// NewSession generates a session with a name and a list of metrics to tag.
// The interval cannot be less than second granularity.
func NewSession(
	metrics []string,
	interval time.Duration,
	pClient *client.Client,
	publisher *wmap.PublishWorkflowMapNode) *Session {

	return &Session{
		Metrics:   metrics,
		Interval:  interval,
		pClient:   pClient,
		Publisher: publisher, // TODO(niklas): Replace with cassandra publisher.
	}
}

// Start an experiment session.
func (s *Session) Start(phaseSession phase.Session) error {
	if s.task != nil {
		return errors.New("task already running")
	}

	// Convert from duration to "Xs" string.
	secondString := fmt.Sprintf("%ds", int(s.Interval.Seconds()))

	t := &task{
		Version: 1,
		Schedule: &client.Schedule{
			Type:     "simple",
			Interval: secondString,
		},
	}

	wf := wmap.NewWorkflowMap()

	for _, metric := range s.Metrics {
		wf.CollectNode.AddMetric(metric, -1)
	}

	// Check if plugins are loaded.
	plugins := NewPlugins(s.pClient)
	loaded, err := plugins.IsLoaded("processor", "session-processor")
	if err != nil {
		return err
	}

	if !loaded {
		// TODO(skonefal): Remove loading this plugin from code.
		goPath := os.Getenv("GOPATH")
		buildPath := path.Join(goPath, "src", "github.com",
			"intelsdi-x", "swan", "build")
		pluginPath := []string{path.Join(
			buildPath, "snap-plugin-processor-session-tagging")}
		err = plugins.Load(pluginPath)
		if err != nil {
			return err
		}
	}

	pr := wmap.NewProcessNode("session-processor", 1)
	pr.AddConfigItem("swan_experiment", phaseSession.ExperimentID)
	pr.AddConfigItem("swan_phase", phaseSession.PhaseID)
	pr.AddConfigItem("swan_repetition", fmt.Sprintf("%d", phaseSession.RepetitionID))
	pr.Add(s.Publisher)
	wf.CollectNode.Add(pr)

	t.Workflow = wf

	r := s.pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true)
	if r.Err != nil {
		return r.Err
	}

	// Save a copy of the task so we can stop it again.
	t.ID = r.ID
	t.State = r.State
	s.task = t

	return nil
}

// Status connects to snap to verify the current state of the task.
func (s *Session) Status() (string, error) {
	if s.task == nil {
		return "", errors.New("snap task not running or not found")
	}

	task := s.pClient.GetTask(s.task.ID)
	if task.Err != nil {
		return "", task.Err
	}

	return task.State, nil
}

// Stop an experiment session.
func (s *Session) Stop() error {
	if s.task == nil {
		return errors.New("snap task not running or not found")
	}

	rs := s.pClient.StopTask(s.task.ID)
	if rs.Err != nil {
		return rs.Err
	}

	rr := s.pClient.RemoveTask(s.task.ID)
	if rr.Err != nil {
		return rr.Err
	}

	s.task = nil

	return nil
}
