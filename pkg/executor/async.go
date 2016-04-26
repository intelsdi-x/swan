package executor

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/command"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Async executor is responsible for orchestrating commander using asynchronus execution.
// It is needed to remove redundant code from Remote & Local Executors
// It runs command as current user.
type Async struct {
	cmdFactory func() command.Command
}

// NewAsync returns a Async instance.
func NewAsync(cmdFactory func() command.Command) Async {
	return Async{cmdFactory}
}

const swanTmpFilesPrefix = "swan_local_executor_"

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (a Async) Execute(command string) (Task, error) {
	log.Debug("Starting ", command)

	// Create temporary output files.
	stdoutFile, err := ioutil.TempFile(os.TempDir(), swanTmpFilesPrefix+"stdout_")
	if err != nil {
		return nil, err
	}
	stderrFile, err := ioutil.TempFile(os.TempDir(), swanTmpFilesPrefix+"stderr_")
	if err != nil {
		return nil, err
	}

	log.Debug("Created temporary files. ",
		"Stdout path:  ", stdoutFile.Name(),
		"Stderr path:  ", stderrFile.Name(),
	)

	cmd := a.cmdFactory()

	cmd.Start(command, stdoutFile, stderrFile)

	log.Debug("Started command.")

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
			// This error happens very rarely and it represent the critical state of the
			// server like volume or HW problems.
			log.Panic("Waiting for task failed. ", err)
		}

		log.Debug(
			"Ended ", command,
			" with output in file: ", stdoutFile.Name(),
			" with err output in file: ", stderrFile.Name(),
			" with status code: ", cmd.ExitCode())
	}()

	return newAsyncTask(cmd, stdoutFile, stderrFile, waitEndChannel), nil
}

const killTimeout = 5 * time.Second

// asyncTask implements Task interface.
type asyncTask struct {
	cmdHandler     command.Command
	stdoutFile     *os.File
	stderrFile     *os.File
	waitEndChannel chan struct{}
}

// newAsyncTask returns a asyncTask instance.
func newAsyncTask(cmdHandler command.Command, stdoutFile *os.File, stderrFile *os.File,
	waitEndChannel chan struct{}) *asyncTask {
	t := &asyncTask{
		cmdHandler:     cmdHandler,
		stdoutFile:     stdoutFile,
		stderrFile:     stderrFile,
		waitEndChannel: waitEndChannel,
	}
	return t
}

// isTerminated checks if waitEndChannel is closed. If it is closed, it means
// that wait ended and task is in terminated state.
func (task *asyncTask) isTerminated() bool {
	select {
	case <-task.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	default:
		return false
	}
}

// Stop terminates the async task.
func (task *asyncTask) Stop() error {
	if task.isTerminated() {
		return nil
	}

	// Sending SIGKILL signal to local task.
	// TODO: Add PID namespace to handle orphan tasks properly.
	err := task.cmdHandler.Kill()
	if err != nil {
		log.Error(err)
		return err
	}

	// Checking if kill was successful.
	isTerminated := task.Wait(killTimeout)
	if !isTerminated {
		return errors.New("Cannot terminate task")
	}

	// No error, task terminated.
	return nil
}

// Status returns a state of the task. If task is terminated it returns the ExitCode as a
// second item in tuple. Otherwise returns nil.
func (task *asyncTask) Status() (TaskState, int) {
	if !task.isTerminated() {
		return RUNNING, -1
	}

	return TERMINATED, task.cmdHandler.ExitCode()
}

// Stdout returns io.Reader to stdout file.
func (task *asyncTask) Stdout() (io.Reader, error) {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return nil, err
	}

	task.stdoutFile.Seek(0, os.SEEK_SET)
	return task.stdoutFile, nil
}

// Stderr returns io.Reader to stderr file.
func (task *asyncTask) Stderr() (io.Reader, error) {
	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return nil, err
	}

	task.stderrFile.Seek(0, os.SEEK_SET)
	return task.stderrFile, nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *asyncTask) Clean() error {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return err
	}

	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return err
	}

	err := task.stdoutFile.Close()
	if err != nil {
		return err
	}

	if err := os.Remove(task.stdoutFile.Name()); err != nil {
		return err
	}

	err = task.stderrFile.Close()
	if err != nil {
		return err
	}

	if err := os.Remove(task.stderrFile.Name()); err != nil {
		return err
	}

	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (task *asyncTask) Wait(timeout time.Duration) bool {
	if task.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case <-task.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}
