package snap

import (
	"fmt"
	"time"

	snapProcessorTag "github.com/intelsdi-x/snap-plugin-processor-tag/tag"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

// SnapteldHTTPEndpoint represents snap daemon address flag.
var SnapteldHTTPEndpoint = conf.NewStringFlag("snapteld_addr", "Snapteld HTTP Endpoint", "http://127.0.0.1:8181")

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
	// TaskName is name of task in Snap.
	TaskName string

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

	// Client to Snapteld.
	pClient *client.Client
}

// NewSession generates a session with a name and a list of metrics to tag.
// The interval cannot be less than second granularity.
func NewSession(
	taskName string,
	metrics []string,
	interval time.Duration,
	pClient *client.Client,
	publisher *wmap.PublishWorkflowMapNode) *Session {

	// Convert from duration to "Xs" string.
	secondString := fmt.Sprintf("%ds", int(interval.Seconds()))

	return &Session{
		TaskName: taskName,
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
func (s *Session) Start(tags string) error {
	if s.task != nil {
		return errors.New("task already running")
	}

	t := &task{
		Name:     s.TaskName,
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

	loaderConfig := DefaultPluginLoaderConfig()
	loaderConfig.SnapteldAddress = s.pClient.URL
	loader, err := NewPluginLoader(loaderConfig)
	if err != nil {
		return err
	}

	if err := loader.Load(TagProcessor); err != nil {
		return err
	}

	pr := wmap.NewProcessNode(snapProcessorTag.Meta().Name, 3)
	pr.AddConfigItem("tags", tags)

	// Add specified publisher to workflow as well.
	pr.Add(s.Publisher)
	wf.CollectNode.Add(pr)

	t.Workflow = wf

	r := s.pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true, 10)
	if r.Err != nil {
		return errors.Wrapf(r.Err, "could not create task %q", t.Name)
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
		return "", errors.New("snap task is not running or not found")
	}

	task := s.pClient.GetTask(s.task.ID)
	if task.Err != nil {
		return "", errors.Wrapf(task.Err, "could not get task name:%q, ID:%q",
			s.task.Name, s.task.ID)
	}

	return task.State, nil
}

// Stop terminates an experiment session and removes Snap task.
// This function blocks until task is stopped.
func (s *Session) Stop() error {
	if s.task == nil {
		return errors.New("snap task not running or not found")
	}

	rs := s.pClient.StopTask(s.task.ID)
	if rs.Err != nil {
		return errors.Wrapf(rs.Err, "could not send stop signal to task %q", s.task.ID)
	}

	err := s.waitForStop()
	if err != nil {
		return errors.Wrapf(err, "could not stop task %q", s.task.ID)
	}

	rr := s.pClient.RemoveTask(s.task.ID)
	if rr.Err != nil {
		return errors.Wrapf(rr.Err, "could not remove task %q", s.task.ID)
	}

	s.task = nil

	return nil
}

// Wait blocks until the task is executed at least once
// (including hits that happened in the past).
func (s *Session) Wait() error {
	for {
		t := s.pClient.GetTask(s.task.ID)
		if t.Err != nil {
			return errors.Wrapf(t.Err, "getting task %q failed", s.task.ID)
		}

		if t.State == "Stopped" || t.State == "Disabled" {
			return errors.Errorf("failed to wait for task: task %q is in state %q (last failure: %q)",
				s.task.ID,
				t.State,
				t.LastFailureMessage)

		}

		if (t.HitCount - (t.FailedCount + t.MissCount)) > 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Session) waitForStop() error {
	for {
		t := s.pClient.GetTask(s.task.ID)
		if t.Err != nil {
			return errors.Wrapf(t.Err, "could not get task %q", s.task.ID)
		}

		if t.State == "Stopped" || t.State == "Disabled" {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}
