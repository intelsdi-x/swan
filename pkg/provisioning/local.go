package provisioning

import (
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"time"
	"syscall"
	"bytes"
	"errors"
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

	s := <-task.statusCh
	task.completeTask(s)
}

// Status gets status of the local task.
func (task LocalTask) Status() Status {
	if (!task.terminated) {
		return Status{code: RunningCode}
	}

	return task.status
}

// Wait blocks until process is terminated or timeout appeared.
// Returns true after timeout exceeds.
func (task *LocalTask) Wait(timeoutMs int) bool {
	if (task.terminated) {
		return false
	}

	if (timeoutMs == 0) {
		s := <-task.statusCh
		task.completeTask(s)

	} else {
		timeoutDuration := time.Duration(timeoutMs) * time.Millisecond

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
// It runs command as current user.
type Local struct{
	isolations []isolation.ProcessIsolation

	// It is important to set additional PGID for parent process and his children
	// to have ability to kill all the children processes.
	setPGID bool
}

// NewLocal returns a Local instance.
func NewLocal(isolations []isolation.ProcessIsolation) Local {
	l := Local{
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

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: l.setPGID}

		// Setting Buffer as io.Writer for Command output.
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Start()

		if (err != nil) {
			panic(err)
		}

		log.Debug("Started with pid ", cmd.Process.Pid)

		// Report the process id.
		taskPidCh <- isolation.TaskPID(cmd.Process.Pid)

		// Wait for task completion.
		cmd.Wait()

		log.Debug(
		  "Ended ", command,
		  " with output: ", stdout.String(),
		  " with err output: ", stderr.String(),
		  " with status code: ",
		  (cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())

		statusCh <- Status{
			(cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(),
			stdout.String(),
			stderr.String(),
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
