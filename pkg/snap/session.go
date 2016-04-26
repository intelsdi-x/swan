package snap

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/nu7hatch/gouuid"
	"strings"
	"time"
)

var (
	pClient *client.Client
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

// Session provides construct for tagging metrics for a specified time span:
// defined by Start() and Stop().
type Session struct {
	// Name is the prefix of the session name
	Name string

	// ID is a unique identifier for the session. Regenerated when Start() is called.
	ID string

	// Interval defines the sample interval for the listed metrics
	Interval time.Duration

	// Metrics to tag in session
	Metrics []string

	// Active task
	task *task

	// Processors
	// Publishers
}

// Lazy connection to snap
func ensureConnected() error {
	if pClient == nil {
		// TODO(niklas): Make 'secure' connection default
		client, err := client.New("http://localhost:8181", "v1", true)

		if err != nil {
			return err
		}

		pClient = client
	}

	return nil
}

// NewSession generates a session with a name and a list of metrics to tag.
// The interval cannot be less than second granularity.
func NewSession(name string, metrics []string, interval time.Duration) (*Session, error) {
	err := ensureConnected()
	if err != nil {
		return nil, err
	}

	return &Session{
		Name:     name,
		Metrics:  metrics,
		Interval: interval,
	}, nil
}

// ListSessions lists current available sessions based on task listing from snap.
func ListSessions() ([]string, error) {
	err := ensureConnected()
	if err != nil {
		return nil, err
	}

	out := []string{}

	tasks := pClient.GetTasks()
	if tasks.Err != nil {
		return out, tasks.Err
	}

	for _, task := range tasks.ScheduledTasks {
		out = append(out, strings.Join([]string{task.ID, task.Name}, "-"))
	}

	return out, nil
}

// Start an experiment session.
// TODO Return session id
func (s *Session) Start() error {
	if s.task != nil {
		return errors.New("task already running")
	}

	err := ensureConnected()
	if err != nil {
		return err
	}

	// Check if plugins are loaded
	// Currently, the mock1 collector, passthru processor and file publisher needs to be loaded.

	// Convert from duration to "Xs" string.
	secondString := fmt.Sprintf("%ds", int(s.Interval.Seconds()))

	t := &task{
		Version: 1,
		Schedule: &client.Schedule{
			Type:     "simple",
			Interval: secondString,
		},
	}

	// Append a UUIDv4 to the session name.
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	s.ID = id.String()

	t.Name = fmt.Sprintf("%s-%s", s.Name, s.ID)

	wf := wmap.NewWorkflowMap()

	// TODO(niklas): Add metrics from session.
	for _, metric := range s.Metrics {
		wf.CollectNode.AddMetric(metric, -1)
	}

	wf.CollectNode.AddConfigItem("/intel/mock/foo", "password", "secret")

	pu := wmap.NewPublishNode("file", 3)
	pu.AddConfigItem("file", "/tmp/swan-snap.out")

	pr := wmap.NewProcessNode("passthru", 1)

	pr.Add(pu)
	wf.CollectNode.Add(pr)

	t.Workflow = wf

	r := pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true)
	if r.Err != nil {
		return r.Err
	}

	t.ID = r.ID
	t.State = r.State

	s.task = t

	return nil
}

// Status connects to snap to verify the current state of the task.
func (s *Session) Status() (string, error) {
	err := ensureConnected()
	if err != nil {
		return "", err
	}

	if s.task == nil {
		return "", errors.New("snap task not running or not found")
	}

	task := pClient.GetTask(s.task.ID)
	if task.Err != nil {
		return "", task.Err
	}

	return task.State, nil
}

// Stop an experiment session.
func (s *Session) Stop() error {
	err := ensureConnected()
	if err != nil {
		return err
	}

	if s.task == nil {
		return errors.New("snap task not running or not found")
	}

	rs := pClient.StopTask(s.task.ID)
	if rs.Err != nil {
		return rs.Err
	}

	rr := pClient.RemoveTask(s.task.ID)
	if rr.Err != nil {
		return rr.Err
	}

	s.task = nil

	return nil
}
