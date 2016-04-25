package snap

import (
  "github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type SnapSession struct {
  // TODO(niklas): Convert to time.Duration
  Interval string

  Plugins []string

  // Collectors
  // Tasks
  // Publishers

  // Store task id
}

func NewSnapSession() (*SnapSession, error) {
  return &SnapSession {}
}

func CanConnectSnapd() (bool, error) {
  return false, nil
}

type task struct {
	Version  int
	Schedule *client.Schedule
	Workflow *wmap.WorkflowMap
	Name     string
	Deadline string
}

func (s* SnapSession) Start(UUID string) error {
  // Check if plugins are loaded

  t := task{
    Version: 1,
    Schedule: &client.Schedule {
      Type: "Simple",
      Interval: SnapSession.Interval,
    },
  }
  t.Name = ctx.String("swan-session")

  wmap := NewWorkflowMap()

  // Populate workflow

  t.Workflow = wmap

	r := pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true)
  if r.Err != nil {
    return r.Err
	}

  return nil
}

func (s* SnapSession) Stop(UUID string) error {
  return nil
}
