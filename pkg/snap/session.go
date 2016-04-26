package snap

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"strings"
  "time"
  "fmt"
)

var (
	pClient *client.Client
)

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

	// Tasks
	// Processors
	// Publishers

	// TODO(niklas): Store task id for Stop()
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
		Name:    name,
		Metrics: metrics,
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

type task struct {
	Version  int
	Schedule *client.Schedule
	Workflow *wmap.WorkflowMap
	Name     string
	Deadline string
}

// Start an experiment session.
// TODO Return session id
func (s *Session) Start() error {
	err := ensureConnected()
	if err != nil {
		return err
	}

	// Check if plugins are loaded

  // Convert from duration to "Xs" string.
  secondString := fmt.Sprintf("%ds", int(s.Interval.Seconds()))

	t := task{
		Version: 1,
		Schedule: &client.Schedule{
			Type:     "simple",
			Interval: secondString,
		},
	}
	t.Name = s.Name

	wf := wmap.NewWorkflowMap()

	// TBD: Populate workflow

	t.Workflow = wf

	r := pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true)
	if r.Err != nil {
		return r.Err
	}

	return nil
}

// Stop an experiment session.
func (s *Session) Stop(UUID string) error {
	err := ensureConnected()
	if err != nil {
		return err
	}

	return nil
}
