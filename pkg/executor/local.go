package executor

import (
	"bytes"
	"errors"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"syscall"
)

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command.
// It runs command as current user.
type Local struct {
}

// NewLocal returns a Local instance.
func NewLocal() Local {
	return Local{}
}

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Execute(command string) (Task, error) {
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

	t := newlocalTask(command, cmd, &stdout, &stdout)

	return t, err
}

// localTask implements Task interface.
type localTask struct {
	command    string
	cmdHandler *exec.Cmd
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	terminated bool
}

// newlocalTask returns a localTask instance.
func newlocalTask(command string, cmdHandler *exec.Cmd,
	stdout *bytes.Buffer, stderr *bytes.Buffer) *localTask {
	t := &localTask{
		command,
		cmdHandler,
		stdout,
		stderr,
		false,
	}
	return t
}

func (task *localTask) completeTask() {
	task.terminated = true
}

func (task localTask) getPid() int {
	return task.cmdHandler.Process.Pid
}

func (task localTask) createStatus() *Status {
	return &Status{
		(task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(),
		task.stdout.String(),
		task.stderr.String(),
	}
}

// Stop terminates the local task.
func (task *localTask) Stop() error {
	if task.terminated {
		return nil
	}

	// We signal the entire process group.
	// The kill syscall interprets a negated PID N as the process group N belongs to.
	log.Debug("Sending SIGTERM to PID ", -task.getPid())
	err := syscall.Kill(-int(task.getPid()), syscall.SIGTERM)
	if err != nil {
		log.Debug(err)
		return err
	}

	// Task should be terminated, however we use timeout to ensure that we don't
	// block stop function in case of error.
	if !WaitWithTimeout(task, 100) {
		// Task is not terminated.
		return errors.New("Task is not yet terminated after kill")
	}

	return err
}

// Status returns a state of the task. If task is terminated it returns the Status as a
// second item in tuple. Otherwise returns nil.
func (task localTask) Status() (TaskState, *Status) {
	if !task.terminated {
		return RUNNING, nil
	}

	return TERMINATED, task.createStatus()
}

// Wait blocks until process is terminated.
func (task *localTask) Wait() {
	if task.terminated {
		return
	}

	// Wait for task completion.
	// NOTE: Wait() returns an error. We grab the process state in any case
	// (success or failure) below, so the error object matters less in the
	// status handling for now.
	err := task.cmdHandler.Wait()
	if err != nil {
		switch err.Error() {
		case "exec: Wait was already called":
			// In case of wait already called we don't need to fill anything yet.
			return
		}
	}

	log.Debug(
		"Ended ", task.command,
		" with output in file: ", task.stdout.String(),
		" with err output in file: ", task.stderr.String(),
		" with status code: ",
		(task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())

	task.completeTask()

	return
}
