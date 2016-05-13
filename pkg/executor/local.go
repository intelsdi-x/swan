package executor

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command.
// It runs command as current user.
type Local struct{}

// NewLocal returns a Local instance.
func NewLocal() Local {
	return Local{}
}

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Execute(command string) (TaskHandle, error) {
	log.Debug("Starting ", command, "' locally ")

	cmd := exec.Command("sh", "-c", command)
	// It is important to set additional Process Group ID for parent process and his children
	// to have ability to kill all the children processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(command, "local")
	if err != nil {
		return nil, err
	}

	log.Debug("Created temporary files ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	log.Debug("Started with pid ", cmd.Process.Pid)

	// Wait End channel is for checking the status of the Wait. If this channel is closed,
	// it means that the wait is completed (either with error or not)
	// This channel will not be used for passing any message.
	waitEndChannel := make(chan struct{})

	// Wait for local task in go routine.
	go func() {
		defer close(waitEndChannel)

		// Wait for task completion.
		// NOTE: Wait() returns an error. We grab the process state in any case
		// (success or failure) below, so the error object matters less in the
		// status handling for now.
		if err := cmd.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				// In case of NON Exit Errors we are not sure if task does
				// terminate so panic.
				// This error happens very rarely and it represent the critical state of the
				// server like volume or HW problems.
				log.Panic("Waiting for local task failed. ", err)
			}
		}

		log.Debug(
			"Ended ", strings.Join(cmd.Args, " "),
			" with output in file: ", stdoutFile.Name(),
			" with err output in file: ", stderrFile.Name(),
			" with status code: ",
			(cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())
	}()

	return newLocalTaskHandle(cmd, stdoutFile, stderrFile, waitEndChannel), nil
}

// localTaskHandle implements TaskHandle interface.
type localTaskHandle struct {
	cmdHandler     *exec.Cmd
	stdoutFile     *os.File
	stderrFile     *os.File
	waitEndChannel chan struct{}
}

// newLocalTaskHandle returns a localTaskHandle instance.
func newLocalTaskHandle(cmdHandler *exec.Cmd, stdoutFile *os.File, stderrFile *os.File,
	waitEndChannel chan struct{}) *localTaskHandle {
	t := &localTaskHandle{
		cmdHandler:     cmdHandler,
		stdoutFile:     stdoutFile,
		stderrFile:     stderrFile,
		waitEndChannel: waitEndChannel,
	}
	return t
}

// isTerminated checks if waitEndChannel is closed. If it is closed, it means
// that wait ended and task is in terminated state.
// NOTE: If it's true then ProcessState is not nil. ProcessState contains information
// about an exited process available after call to Wait or Run.
func (taskHandle *localTaskHandle) isTerminated() bool {
	select {
	case <-taskHandle.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	default:
		return false
	}
}

func (taskHandle *localTaskHandle) getPid() int {
	return taskHandle.cmdHandler.Process.Pid
}

// Stop terminates the local task.
func (taskHandle *localTaskHandle) Stop() error {
	if taskHandle.isTerminated() {
		return nil
	}

	// Sending SIGKILL signal to local task.
	// TODO: Add PID namespace to handle orphan tasks properly.
	log.Debug("Sending ", syscall.SIGKILL, " to PID ", -taskHandle.getPid())
	err := syscall.Kill(-taskHandle.getPid(), syscall.SIGKILL)
	if err != nil {
		log.Error(err)
		return err
	}

	// Checking if kill was successful.
	isTerminated := taskHandle.Wait(killTimeout)
	if !isTerminated {
		return errors.New("Cannot terminate task")
	}

	// No error, task terminated.
	return nil
}

// Status returns a state of the task.
func (taskHandle *localTaskHandle) Status() TaskState {
	if !taskHandle.isTerminated() {
		return RUNNING
	}

	return TERMINATED
}

// ExitCode returns a exitCode. If task is not terminated it returns error.
func (taskHandle *localTaskHandle) ExitCode() (int, error) {
	if !taskHandle.isTerminated() {
		return -1, errors.New("Task is not terminated")
	}

	return (taskHandle.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(), nil
}

// StdoutFile returns a file handle for file to the task's stdout file.
func (taskHandle *localTaskHandle) StdoutFile() (*os.File, error) {
	if _, err := os.Stat(taskHandle.stdoutFile.Name()); err != nil {
		return nil, err
	}

	taskHandle.stdoutFile.Seek(0, os.SEEK_SET)
	return taskHandle.stdoutFile, nil
}

// StderrFile returns a file handle for file to the task's stderr file.
func (taskHandle *localTaskHandle) StderrFile() (*os.File, error) {
	if _, err := os.Stat(taskHandle.stderrFile.Name()); err != nil {
		return nil, err
	}

	taskHandle.stderrFile.Seek(0, os.SEEK_SET)
	return taskHandle.stderrFile, nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (taskHandle *localTaskHandle) Clean() error {
	// Close stdout.
	stdoutErr := taskHandle.stdoutFile.Close()

	// Close stderr.
	stderrErr := taskHandle.stderrFile.Close()

	if stdoutErr != nil {
		return stdoutErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (taskHandle *localTaskHandle) EraseOutput() error {
	// Remove stdout file.
	if err := os.Remove(taskHandle.stdoutFile.Name()); err != nil {
		return err
	}

	// Remove stderr file.
	if err := os.Remove(taskHandle.stderrFile.Name()); err != nil {
		return err
	}

	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (taskHandle *localTaskHandle) Wait(timeout time.Duration) bool {
	if taskHandle.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case <-taskHandle.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}
