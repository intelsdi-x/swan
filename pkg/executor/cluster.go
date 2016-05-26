package executor

import (
	"github.com/gocql/gocql"
	"github.com/satori/go.uuid"
	"time"
)

// Cluster ...
type Cluster struct {
	Name    string
	session *gocql.Session
}

// Job ...
type Job struct {
	Id           gocql.UUID
	Name         string
	CurrentState string
	DesiredState string
	AssignedHost string
	Command      string
  Restart      bool
}

var (
	createKeyspaceCQL = "CREATE KEYSPACE IF NOT EXISTS swan WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};"
	createTableCQL    = "CREATE TABLE IF NOT EXISTS swan.jobs (id uuid, added timestamp, name text, current_state text, desired_state text, assigned_host text, command text, PRIMARY KEY (id));"
	addJobCQL         = "INSERT INTO swan.jobs (id, added, name, current_state, desired_state, assigned_host, command) VALUES (?, ?, ?, ?, ?, ? ,?)"
	unassignedJobsCQL = "SELECT id, name, current_state, desired_state, assigned_host, command FROM swan.jobs where assigned_host = '' LIMIT 1 allow filtering"
	updateHostCQL     = "UPDATE swan.jobs SET assigned_host=?, current_state=? WHERE id=? IF assigned_host = ''"
	updateStateCQL    = "UPDATE swan.jobs SET current_state=? WHERE id=? IF current_state=?"
)

func NewCluster(name string, coordinator string) (Executor, error) {
	cluster := gocql.NewCluster(coordinator)
	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 4

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	if err := session.Query(createKeyspaceCQL).Exec(); err != nil {
		return nil, err
	}

	if err := session.Query(createTableCQL).Exec(); err != nil {
		return nil, err
	}

	return &Cluster{
		session: session,
	}, nil
}

func (c *Cluster) Execute(command string) (TaskHandle, error) {
	if err := c.session.Query(addJobCQL, uuid.NewV4().String(), time.Now().UnixNano()/1000, c.Name, "unscheduled", "started", "", command).Exec(); err != nil {
		return nil, err
	}

	return nil, nil
}

// Agent ...
type Agent struct {
	Name       string
	session    *gocql.Session
	CurrentJob *Job
}

func NewAgent(name string, coordinator string) (*Agent, error) {
	cluster := gocql.NewCluster(coordinator)
	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 4

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &Agent{
		Name:       name,
		session:    session,
		CurrentJob: nil,
	}, nil
}

func (a *Agent) StealJob() error {
	j := &Job{}

	iter := a.session.Query(unassignedJobsCQL).Iter()

	iter.Scan(&j.Id, &j.Name, &j.CurrentState, &j.DesiredState, &j.AssignedHost, &j.Command)

	if err := iter.Close(); err != nil {
		return err
	}

	if err := a.session.Query(updateHostCQL, a.Name, "scheduled", j.Id).Exec(); err != nil {
		return err
	}

	a.CurrentJob = j
	return nil
}

func (a *Agent) Reconcile() {
}

func (a *Agent) startJob() {
  if err := a.session.Query(updateStateCQL, "starting", "scheduled", j.Id).Exec(); err != nil {
		return err
	}

  // Fork/exec
  if err := a.session.Query(updateStateCQL, "running", "starting", j.Id).Exec(); err != nil {
		return err
	}
}

func (a *Agent) terminateJob() {
  if err := a.session.Query(updateStateCQL, "terminated", "running", j.Id).Exec(); err != nil {
		return err
	}
}
