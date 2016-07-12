package executor

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"fmt"
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

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Execute(command string) (TaskHandle, error) {
	log.Debug("Starting ", l.commandDecorators.Decorate(command), "' locally ")

	cmd := exec.Command("sh", "-c", l.commandDecorators.Decorate(command))

	// TODO: delete this as we use PID namespace instead
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(command, "local")
	if err != nil {
		return nil, errors.Wrapf(err, "createExecutorOutputFiles for command %q failed", command)
	}

	log.Debug("Created temporary files ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrapf(err, "command %q start failed", command)
	}

	log.Debug("Started with pid ", cmd.Process.Pid)

	// Wait End channel is for checking the status of the Wait. If this channel is closed,
	// it means that the wait is completed (either with error or not)
	// This channel will not be used for passing any message.
	hasProcessExited := make(chan struct{})
	hasStopOrWaitInvoked := make(chan struct{})

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

		// Flush write buffers to disk to prevent a race with the caller.
		stdoutFile.Sync()
		stderrFile.Sync()

		pid := cmd.Process.Pid
		select {
		// If Wait or Stop has been invoked on TaskHandle, then exit is expected.
		case <-hasStopOrWaitInvoked:
			// logrus escapes newline, making this log unreadable otherwise
			log.Debugf("%d Process %q ended\n", pid, strings.Join(cmd.Args, " "))
			log.Debugf("%d Stdout stored in %q", pid, stdoutFile.Name())
			log.Debugf("%d Stderr stored in %q", pid, stderrFile.Name())
			log.Debugf("%d Exit code: %d", pid, (cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())
		default:
			// If process exited before Wait or Stop, it might have ended prematurely.
			stdoutTail, err := readTail(stdoutFile.Name())
			if err != nil {
				stdoutTail = fmt.Sprintf("%v", err)
			}
			stderrTail, err := readTail(stderrFile.Name())
			if err != nil {
				stderrTail = fmt.Sprintf("%v", err)
			}

			log.Errorf("%d Process %q might have ended prematurely", pid, strings.Join(cmd.Args, " "))
			log.Errorf("%d Stdout stored in %q", pid, stdoutFile.Name())
			log.Errorf("%d Stderr stored in %q", pid, stderrFile.Name())
			log.Errorf("%d Exit code: %d", pid, (cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())
			log.Errorf("%d Last 10 lines of stdout:", pid)
			log.Errorf("%d %q", pid, stdoutTail)
			log.Errorf("%d Last 10 lines of stderr:", pid)
			log.Errorf("%d %q", pid, stderrTail)
		}
	}()

	return newLocalTaskHandle(cmd, stdoutFile, stderrFile, hasProcessExited, hasStopOrWaitInvoked), nil
}

// localTaskHandle implements TaskHandle interface.
type localTaskHandle struct {
	cmdHandler *exec.Cmd
	stdoutFile *os.File
	stderrFile *os.File

	// This channel is closed immedietaly when process exits.
	// It is used to signal task termination.
	processHasExited chan struct{}

	// This channel is closed when Stop or Wait has been invoked on TaskHandle.
	// It is used to signal that process exit is expected by user.
	hasStopOrWaitInvoked chan struct{}
	// internal flag controlling closing of hasStopOrWaitInvoked channel
	stopOrWaitChannelClosed bool
}

// newLocalTaskHandle returns a localTaskHandle instance.
func newLocalTaskHandle(
	cmdHandler *exec.Cmd,
	stdoutFile *os.File,
	stderrFile *os.File,
	processHasExited chan struct{},
	hasStopOrWaitBeenInvoked chan struct{}) *localTaskHandle {
	t := &localTaskHandle{
		cmdHandler:           cmdHandler,
		stdoutFile:           stdoutFile,
		stderrFile:           stderrFile,
		processHasExited:     processHasExited,
		hasStopOrWaitInvoked: hasStopOrWaitBeenInvoked,
	}
	return t
}

// isTerminated checks if channel processHasExited is closed. If it is closed, it means
// that wait ended and task is in terminated state.
// NOTE: If it's true then ProcessState is not nil. ProcessState contains information
// about an exited process available after call to Wait or Run.
func (taskHandle *localTaskHandle) isTerminated() bool {
	select {
	case <-taskHandle.processHasExited:
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
	taskHandle.stopOrWaitInvoked()
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
	if _, err := os.Stat(taskHandle.stdoutFile.Name()); err != nil {
		return nil, errors.Wrapf(err, "os.stat on file %q failed", taskHandle.stdoutFile.Name())
	}

	taskHandle.stdoutFile.Seek(0, os.SEEK_SET)
	return taskHandle.stdoutFile, nil
}

// StderrFile returns a file handle for file to the task's stderr file.
func (taskHandle *localTaskHandle) StderrFile() (*os.File, error) {
	if _, err := os.Stat(taskHandle.stderrFile.Name()); err != nil {
		return nil, errors.Wrapf(err, "os.stat on file %q failed", taskHandle.stderrFile.Name())
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
		return errors.Wrapf(stdoutErr, "close on file %q failed", taskHandle.stdoutFile.Name())
	}

	if stderrErr != nil {
		return errors.Wrapf(stderrErr, "close on file %q failed", taskHandle.stderrFile.Name())
	}

	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (taskHandle *localTaskHandle) EraseOutput() error {
	outputDir, _ := path.Split(taskHandle.stdoutFile.Name())

	// Remove temporary directory created for stdout and stderr.
	if err := os.RemoveAll(outputDir); err != nil {
		return errors.Wrapf(err, "os.RemoveAll of directory %q failed", outputDir)
	}
	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (taskHandle *localTaskHandle) Wait(timeout time.Duration) bool {
	taskHandle.stopOrWaitInvoked()
	if taskHandle.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case <-taskHandle.processHasExited:
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

func (taskHandle *localTaskHandle) stopOrWaitInvoked() {
	if taskHandle.stopOrWaitChannelClosed {
		return
	}

	close(taskHandle.hasStopOrWaitInvoked)
	taskHandle.stopOrWaitChannelClosed = true
}
