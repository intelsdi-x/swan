package snap

import (
	"errors"
	"fmt"
	snapProcessorTag "github.com/intelsdi-x/snap-plugin-processor-tag/tag"
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
	// Schedule defines the schedule type and interval for the listed metrics.
	Schedule *client.Schedule

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

	// Convert from duration to "Xs" string.
	secondString := fmt.Sprintf("%ds", int(interval.Seconds()))

	return &Session{
		Schedule: &client.Schedule{
			Type:     "simple",
			Interval: secondString,
		},
		Metrics:                metrics,
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

	t := &task{
		Version:  1,
		Schedule: s.Schedule,
	}

	wf := wmap.NewWorkflowMap()

	for _, metric := range s.Metrics {
		wf.CollectNode.AddMetric(metric, -1)
	}

	for _, configItem := range s.CollectNodeConfigItems {
		wf.CollectNode.AddConfigItem(configItem.Ns, configItem.Key, configItem.Value)
	}

	// Check if `processor-tag` plugins is loaded.
	// Get the Name of the processor-tag plugin from its Meta.
	plugins := NewPlugins(s.pClient)
	loaded, err := plugins.IsLoaded("processor", snapProcessorTag.Meta().Name)
	if err != nil {
		return err
	}

	if !loaded {
		goPath := os.Getenv("GOPATH")
		pluginPath := []string{path.Join(
			goPath, "bin", "snap-plugin-processor-tag")}
		err = plugins.Load(pluginPath)
		if err != nil {
			return err
		}
	}

	pr := wmap.NewProcessNode(snapProcessorTag.Meta().Name, 3)

	// Constructing Tags config item as stated in
	// https://github.com/intelsdi-x/snap-plugin-processor-tag/README.md
	pr.AddConfigItem("tags", fmt.Sprintf("swan_experiment:%s,swan_phase:%s,swan_repetition:%d",
		phaseSession.ExperimentID,
		phaseSession.PhaseID,
		phaseSession.RepetitionID,
	))

	// Add specified publisher to workflow as well.
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

// IsRunning checks if Snap task is running.
func (s *Session) IsRunning() bool {
	status, err := s.status()
	if err != nil {
		return false
	}
	return status == "Running"
}

// Status connects to snap to verify the current state of the task.
func (s *Session) status() (string, error) {
	if s.task == nil {
		return "", errors.New("snap task not running or not found")
	}

	task := s.pClient.GetTask(s.task.ID)
	if task.Err != nil {
		return "", task.Err
	}

	return task.State, nil
}

// StopAndRemove terminates an experiment session and removes Snap task.
// This function blocks until task is stopped.
func (s *Session) StopAndRemove() error {
	if s.task == nil {
		return errors.New("snap task not running or not found")
	}

	rs := s.pClient.StopTask(s.task.ID)
	if rs.Err != nil {
		return rs.Err
	}
	s.waitForStop()

	rr := s.pClient.RemoveTask(s.task.ID)
	if rr.Err != nil {
		return rr.Err
	}

	s.task = nil

	return nil
}

// WaitForSuccessfulHit blocks until the task is executed at least once
// (including hits that happened in the past).
func (s *Session) WaitForSuccessfulHit() {
	for {
		t := s.pClient.GetTask(s.task.ID)
		if (t.HitCount - (t.FailedCount + t.MissCount)) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Session) waitForStop() {
	for {
		t := s.pClient.GetTask(s.task.ID)
		if t.State == "Stopped" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
