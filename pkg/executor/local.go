package executor

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type taskPID int64

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
	statusChannel := make(chan Status)
	stdoutDir, err := ioutil.TempFile("/tmp/", "stdout")
	if err != nil {
		return nil, err
	}
	stderrDir, err := ioutil.TempFile("/tmp/", "stderr")
	if err != nil {
		return nil, err
	}

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

	err = cmd.Start()
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

		var exitCode int
		// If Process exited on his own, show the exitStatus.
		if (cmd.ProcessState.Sys().(syscall.WaitStatus)).Exited() {
			exitCode = (cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus()
		} else {
			// Show what signal caused the termination.
			exitCode = -int((cmd.ProcessState.Sys().(syscall.WaitStatus)).Signal())
		}

		log.Debug(
			"Ended ", command,
			" with output: ", stdout.String(),
			" with err output: ", stderr.String(),
			" with status code: ", exitCode)

		ioutil.WriteFile(stdoutDir.Name(), stdout.Bytes(), 600)
		ioutil.WriteFile(stderrDir.Name(), stderr.Bytes(), 600)

		statusChannel <- Status{
			&exitCode,
		}
	}()

	taskPid := taskPID(cmd.Process.Pid)

	t := newlocalTask(taskPid, statusChannel, stdoutDir.Name(), stderrDir.Name())

	return t, err
}

// Stdout returns io.Reader to stdout file.
func (task *localTask) Stdout() (io.Reader, error) {
	if _, err := os.Stat(task.stdoutDir); err != nil {
		return nil, err
	}
	stdoutFile, err := os.Open(task.stdoutDir)
	if err != nil {
		return nil, err
	}
	return io.Reader(stdoutFile), nil
}

// Stderr returns io.Reader to stderr file.
func (task *localTask) Stderr() (io.Reader, error) {
	if _, err := os.Stat(task.stderrDir); err != nil {
		return nil, err
	}
	stderrFile, err := os.Open(task.stderrDir)
	if err != nil {
		return nil, err
	}
	return io.Reader(stderrFile), nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *localTask) Clean() error {
	if _, err := os.Stat(task.stdoutDir); err != nil {
		return err
	}
	if _, err := os.Stat(task.stderrDir); err != nil {
		return err
	}
	if err := os.Remove(task.stdoutDir); err != nil {
		return err
	}
	if err := os.Remove(task.stderrDir); err != nil {
		return err
	}
	return nil
}

func (task *localTask) GetStdoutDir() (string, error) {
	if _, err := os.Stat(task.stdoutDir); err != nil {
		return "", err
	}
	return task.stdoutDir, nil
}

func (task *localTask) GetStderrDir() (string, error) {
	if _, err := os.Stat(task.stderrDir); err != nil {
		return "", err
	}
	return task.stderrDir, nil
}

// localTask implements Task interface.
type localTask struct {
	pid           taskPID
	statusChannel chan Status
	status        Status
	terminated    bool
	stdoutDir     string
	stderrDir     string
}

// newlocalTask returns a localTask instance.
func newlocalTask(pid taskPID, statusChannel chan Status, stdoutDir string, stderrDir string) *localTask {
	t := &localTask{
		pid,
		statusChannel,
		Status{},
		false,
		stdoutDir,
		stderrDir,
	}
	return t
}

func (task *localTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusChannel = nil
}

// Stop terminates the local task.
func (task *localTask) Stop() error {
	if task.terminated {
		return nil
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

// Status returns a state of the task. If task is terminated it returns the Status as a
// second item in tuple. Otherwise returns nil.
func (task localTask) Status() (TaskState, *int) {
	if !task.terminated {
		return RUNNING, nil
	}

	return TERMINATED, task.status.ExitCode
}

// Wait blocks until process is terminated or timeout appeared.
// Returns true when process terminates before timeout, otherwise false.
func (task *localTask) Wait(timeoutMs int) bool {
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
