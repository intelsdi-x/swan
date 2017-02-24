package executor

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pkg/errors"
)

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command.
// It runs command as current user.
type Local struct {
	commandDecorators isolation.Decorator
}

// NewLocal returns instance of local executors without any isolators.
func NewLocal() Local {
	return NewLocalIsolated(isolation.Decorators{})
}

// NewLocalIsolated returns a Local instance with some isolators set.
func NewLocalIsolated(decorator isolation.Decorator) Local {
	return Local{commandDecorators: decorator}
}

// Name returns user-friendly name of executor.
func (l Local) Name() string {
	return "Local Executor"
}

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Execute(command string) (TaskHandle, error) {
	log.Debug("Starting '", l.commandDecorators.Decorate(command), "' locally ")

	cmd := exec.Command("sh", "-c", l.commandDecorators.Decorate(command))

	// TODO: delete this as we use PID namespace instead
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	outputDirectory, err := createOutputDirectory(command, "local")
	if err != nil {
		return nil, errors.Wrapf(err, "createOutputDirectory for command %q failed", command)
	}
	stdoutFile, stderrFile, err := createExecutorOutputFiles(outputDirectory)
	if err != nil {
		removeDirectory(outputDirectory)
		return nil, errors.Wrapf(err, "createExecutorOutputFiles for command %q failed", command)
	}

	log.Debug("Created temporary files ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	err = cmd.Start()
	if err != nil {
		removeDirectory(outputDirectory)
		return nil, errors.Wrapf(err, "command %q start failed", command)
	}

	log.Debug("Started with pid ", cmd.Process.Pid)

	// hasProcessExited channel is closed when launched process exits.
	hasProcessExited := make(chan struct{})

	taskHandle := newLocalTaskHandle(cmd, stdoutFile.Name(), stderrFile.Name(), hasProcessExited)

	// Wait for local task in go routine.
	go func() {
		defer close(hasProcessExited)

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
				err = errors.Wrap(err, "wait returned with NON exit error")
				log.Panicf("Waiting for local task failed\n%+v", err)
			}
		}

		err = syncAndClose(stdoutFile)
		if err != nil {
			log.Errorf("Cannot syncAndClose stdout file: %s", err.Error())
		}
		err = syncAndClose(stderrFile)
		if err != nil {
			log.Errorf("Cannot syncAndClose stderrFile file: %s", err.Error())
		}
	}()

	return taskHandle, nil
}

// localTaskHandle implements TaskHandle interface.
type localTaskHandle struct {
	cmdHandler     *exec.Cmd
	stdoutFilePath string
	stderrFilePath string

	// This channel is closed immediately when process exits.
	// It is used to signal task termination.
	hasProcessExited chan struct{}

	// internal flag controlling closing of hasStopOrWaitInvoked channel
	stopOrWaitChannelClosed bool
}

// newLocalTaskHandle returns a localTaskHandle instance.
func newLocalTaskHandle(
	cmdHandler *exec.Cmd,
	stdoutFile string,
	stderrFile string,
	hasProcessExited chan struct{}) *localTaskHandle {
	t := &localTaskHandle{
		cmdHandler:       cmdHandler,
		stdoutFilePath:   stdoutFile,
		stderrFilePath:   stderrFile,
		hasProcessExited: hasProcessExited,
	}
	return t
}

// isTerminated checks if channel processHasExited is closed. If it is closed, it means
// that wait ended and task is in terminated state.
// NOTE: If it's true then ProcessState is not nil. ProcessState contains information
// about an exited process available after call to Wait or Run.
func (taskHandle *localTaskHandle) isTerminated() bool {
	select {
	case <-taskHandle.hasProcessExited:
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
		return errors.Wrapf(err, "kill of PID %d of application %q failed",
			-taskHandle.getPid(), taskHandle.cmdHandler.Path)
	}

	// Checking if kill was successful.
	isTerminated := taskHandle.Wait(killTimeout)
	if !isTerminated {
		return errors.Errorf("cannot terminate running %q application",
			taskHandle.cmdHandler.Path)
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
		return -1, errors.Errorf("task %q is not terminated", taskHandle.cmdHandler.Path)
	}

	return (taskHandle.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(), nil
}

// StdoutFile returns a file handle for file to the task's stdout file.
func (taskHandle *localTaskHandle) StdoutFile() (*os.File, error) {
	return openFile(taskHandle.stdoutFilePath)
}

// StderrFile returns a file handle for file to the task's stderr file.
func (taskHandle *localTaskHandle) StderrFile() (*os.File, error) {
	return openFile(taskHandle.stderrFilePath)
}

// Deprecated: Does nothing.
func (taskHandle *localTaskHandle) Clean() error {
	return nil
}

// EraseOutput deletes the directory where stdout file resides.
func (taskHandle *localTaskHandle) EraseOutput() error {
	outputDir := filepath.Dir(taskHandle.stdoutFilePath)
	return removeDirectory(outputDir)
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
	case <-taskHandle.hasProcessExited:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}

func (taskHandle *localTaskHandle) Address() string {
	return "127.0.0.1"
}
