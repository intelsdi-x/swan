package provisioning

import (
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"time"
)

// LocalTask implements Task interface.
type LocalTask struct{
	pid isolation.TaskPID
	statusCh chan Status
	status Status

	completed bool
}

// NewLocalTask returns a LocalTask instance.
func NewLocalTask(pid isolation.TaskPID, statusCh chan Status) *LocalTask {
	t := &LocalTask{
		pid,
		statusCh,
		Status{},
		false,
	}
	return t
}

func (task *LocalTask) completeTask(status Status) {
	task.completed = true
	task.status = status
	task.statusCh = nil
}

func (task *LocalTask) Stop() error {
	if (task.completed) {
		return NewError("Task is not running.")
	}

	// TODO(bp): Stop pid with.
	return NewError("Not implemented")
}

func (task LocalTask) Status() Status {
	// TODO(bp): Get status.
	return task.status
}

func (task *LocalTask) Wait(timeoutSeconds int) error {
	if (task.completed) {
		return nil
	}

	if (timeoutSeconds == 0) {
		s := <-task.statusCh
		task.completeTask(s)

	} else {
		timeoutDuration := time.Duration(timeoutSeconds) * time.Second

		select {
		case s := <-task.statusCh:
			task.completeTask(s)
		case <-time.After(timeoutDuration):
			return NewError("Timeout occured after: ", timeoutDuration)
		}
	}

	return nil
}

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command. It also needed to setup given isolation
// using Isolation Manager.
type Local struct{
	user string
	isolations []isolation.Isolation
}

// NewLocal returns a Local instance.
func NewLocal(user string, isolations []isolation.Isolation) Local {
	l := Local{
		user: user,
		isolations: isolations,
	}
	return l
}


// Run runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (l Local) Run(command string) (Task, error) {
	statusCh := make(chan Status)

	taskPidCh := make(chan isolation.TaskPID)

	// Do initialization of the isolation synchronously.
	for _, isolation := range l.isolations {
		isolation.Init()
	}

	// Run task in local locally.
	go func() {
		log.Debug("Starting ", command)

		cmd := exec.Command("sh", "-c", command)
		err := cmd.Start()
		if (err != nil) {
			panic(err)
		}

		log.Debug("Started with pid ", cmd.Process.Pid)

		// Report the process id.
		taskPidCh <- isolation.TaskPID(cmd.Process.Pid)
		// Wait for task completion.
		cmd.Wait()

		log.Debug("Ended ", command)

		// TODO(bplotka): Fetch status code.

		output, _ := cmd.Output()
		statusCh <- Status{0, string(output)}
	}()

	// Get PID
	taskPid := <-taskPidCh

	// Perform rest of the isolation synchronously.
	for _, isolation := range l.isolations {
		isolation.Perform(taskPid)
	}

	t := NewLocalTask(taskPid, statusCh)

	return t, nil
}