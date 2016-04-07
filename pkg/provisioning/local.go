package provisioning

import (
	"bytes"
	"errors"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"syscall"
	"time"
)

type TaskPID int64

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command.
// It runs command as current user.
type Local struct {
}

// NewLocal returns a Local instance.
func NewLocal() Local {
	return Local{}
}

// Run runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Run(command string) (Task, error) {
	statusChannel := make(chan Status)

	log.Debug("Starting ", command)

	cmd := exec.Command("sh", "-c", command)

	// It is important to set additional Process Group ID for parent process and his children
	// to have ability to kill all the children processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Setting Buffer as io.Writer for Command output.
	// TODO: Write to temporary files instead of keeping in memory.
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	log.Debug("Started with pid ", cmd.Process.Pid)

	// Wait for local task in goroutine.
	go func() {
		// Wait for task completion.
		// NOTE: Wait() returns an error. We grab the process state in any case
		// (success or failure) below, so the error object matters less in the
		// status handling for now.
		cmd.Wait()

		log.Debug(
			"Ended ", command,
			" with output: ", stdout.String(),
			" with err output: ", stderr.String(),
			" with status code: ",
			(cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())

		statusChannel <- Status{
			(cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(),
			stdout.String(),
			stderr.String(),
		}
	}()

	taskPid := TaskPID(cmd.Process.Pid)

	t := newLocalTask(taskPid, statusChannel)

	return t, err
}

// LocalTask implements Task interface.
type LocalTask struct {
	pid           TaskPID
	statusChannel chan Status
	status        Status
	terminated    bool
}

// newLocalTask returns a LocalTask instance.
func newLocalTask(pid TaskPID, statusChannel chan Status) *LocalTask {
	t := &LocalTask{
		pid,
		statusChannel,
		Status{},
		false,
	}
	return t
}

func (task *LocalTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusChannel = nil
}

// Stop terminates the local task.
func (task *LocalTask) Stop() error {
	if task.terminated {
		return errors.New("Task is not running.")
	}

	// We signal the entire process group.
	// The kill syscall interprets a negated PID N as the process group N belongs to.
	log.Debug("Sending SIGTERM to PID ", -task.pid)
	err := syscall.Kill(-int(task.pid), syscall.SIGTERM)
	if err != nil {
		return err
	}

	s := <-task.statusChannel
	task.completeTask(s)

	return err
}

// Status gets status of the local task.
func (task LocalTask) Status() Status {
	if !task.terminated {
		return Status{code: 0}
	}

	return task.status
}

// Wait blocks until process is terminated or timeout appeared.
// Returns true when process terminates before timeout, otherwise false.
func (task *LocalTask) Wait(timeoutMs int) bool {
	if task.terminated {
		return true
	}

	if timeoutMs == 0 {
		s := <-task.statusChannel
		task.completeTask(s)
		return true
	}

	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	result := true

	select {
	case s := <-task.statusChannel:
		task.completeTask(s)
	case <-time.After(timeoutDuration):
		result = false
	}

	return result
}
