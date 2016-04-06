package provisioning

import (
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"time"
	"syscall"
)

// LocalTask implements Task interface.
type LocalTask struct{
	pid isolation.TaskPID
	statusCh chan Status
	status Status
	terminated bool
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
	task.terminated = true
	task.status = status
	task.statusCh = nil
}

// Stop terminates the local task.
func (task *LocalTask) Stop() {
	if (task.terminated) {
		panic("Task is not running.")
	}

	log.Debug("Sending SIGTERM to PID ", -task.pid)
	err := syscall.Kill(-int(task.pid), syscall.SIGTERM)
	if (err != nil) {
		panic(err)
	}
}

// Status gets status of the local task.
func (task LocalTask) Status() Status {
	// TODO(bp): Get status.
	return task.status
}

// Wait blocks until process is terminated or timeout appeared.
// Returns true after timeout exceeds.
func (task *LocalTask) Wait(timeoutSeconds int) bool {
	if (task.terminated) {
		return false
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
			return true
		}
	}

	return false
}

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command. It also needed to setup given isolation
// using Isolation Manager.
type Local struct{
	user string
	isolations []isolation.Isolation

	// It is important to set additional PGID for parent process and his children
	// to have ability to kill all the children processes.
	setPGID bool
}

// NewLocal returns a Local instance.
func NewLocal(user string, isolations []isolation.Isolation) Local {
	l := Local{
		user: user,
		isolations: isolations,
		setPGID: true,
	}
	return l
}


// Run runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (l Local) Run(command string) (Task) {
	statusCh := make(chan Status)

	taskPidCh := make(chan isolation.TaskPID)

	// Run task in local locally.
	go func() {
		log.Debug("Starting ", command)

		cmd := exec.Command("sh", "-c", command)

		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Setpgid = l.setPGID

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
		statusCh <- Status{
			cmd.ProcessState.Sys().(syscall.WaitStatus),
			string(output),
		}
	}()

	// Get PID.
	taskPid := <-taskPidCh

	// Perform rest of the isolation synchronously.
	for _, isolation := range l.isolations {
		isolation.Isolate(taskPid)
	}

	t := NewLocalTask(taskPid, statusCh)

	return t
}
