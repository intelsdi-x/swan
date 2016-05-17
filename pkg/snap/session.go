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

// CollectNodeConfigItem represents ConfigItem in CollectWorkflowMapNode.
type CollectNodeConfigItem struct {
	Ns    string
	Key   string
	Value interface{}
}

// Session provides construct for tagging metrics for a specified time span
// defined by Start() and Stop().
type Session struct {
	// Interval defines the sample interval for the listed metrics.
	Interval time.Duration

	// Metrics to tag in session.
	Metrics []string

	// CollectNodeConfigItems represent ConfigItems for CollectNode.
	CollectNodeConfigItems []CollectNodeConfigItem

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
		Metrics:                metrics,
		Interval:               interval,
		pClient:                pClient,
		Publisher:              publisher, // TODO(niklas): Replace with cassandra publisher.
		CollectNodeConfigItems: []CollectNodeConfigItem{},
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

	for _, configItem := range s.CollectNodeConfigItems {
		wf.CollectNode.AddConfigItem(configItem.Ns, configItem.Key, configItem.Value)
	}

	// Check if session processor plugins is loaded.
	plugins := NewPlugins(s.pClient)
	loaded, err := plugins.IsLoaded("processor", "session-processor")
	if err != nil {
		return err
	}

	if !loaded {
		// TODO(skonefal): Remove loading this plugin from code.
		// NOTE(bp): Disagree with above (:
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

	// TODO(bp): Make sure Test completed it's work. (!IMPORTANT)
	// Hint from @Iwan: Look on to achieve that.
	// https://github.com/intelsdi-x/swan/blob/iwan/sprint-17-demo/misc/snap-plugin-collector-mutilate/snap-with-mutilate/test.sh
	// It is convenient to just check how many successful task runs has been.
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
