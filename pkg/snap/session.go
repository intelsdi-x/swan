package snap

import (
  "github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type SnapSession struct {
  Name string

  ID string

  // TODO(niklas): Convert to time.Duration
  Interval string

  Plugins []string

  pClient    *client.Client

  // Collectors
  // Tasks
  // Publishers

  // Store task id
}

func NewSnapSession(name string) (*SnapSession, error) {
  // TODO(niklas): Make 'secure' connection default
  pClient, err := client.New("http://localhost:8181", "v1", true)
  if err != nil {
    return nil, err
  }

  return &SnapSession {
    Name: name,
    pClient: pClient,
  }, nil
}

func CanConnectSnapd() (bool, error) {
  return false, nil
}

func ListSessions() ([]string, error) {
  return []string{}, nil
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
      Interval: s.Interval,
    },
  }
  t.Name = s.Name

  wf := wmap.NewWorkflowMap()

  // TBD: Populate workflow

  t.Workflow = wf

	r := s.pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, true)
  if r.Err != nil {
    return r.Err
	}

  return nil
}

func (s* SnapSession) Stop(UUID string) error {
  return nil
}
