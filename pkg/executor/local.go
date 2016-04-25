package executor

import (
	"bytes"
	"errors"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
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

	log.Debug("Started with PID ", cmd.Process.Pid)

	// Wait Err channel is for gethering error potential error messages from cmd.Wait function.
	waitErrChannel := make(chan error, 1)
	// Wait End channel is for checking the status of the Wait. If this channel is closed,
	// it means that the wait is completed (either with error or not)
	// This channel will not be used for passing any message.
	waitEndChannel := make(chan struct{})

	// Wait for local task in goroutine.
	go func() {
		// Wait for task completion.
		// NOTE: Wait() returns an error. We grab the process state in any case
		// (success or failure) below, so the error object matters less in the
		// status handling for now.
		if err := cmd.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				// In case of NON Exit Errors we are not sure if task does
				// terminate so return error.
				log.Error("Waiting for task failed. ", err)
				waitErrChannel <- err

				close(waitErrChannel)
				close(waitEndChannel)
				return
			}
		}

		log.Debug(
			"Ended ", strings.Join(cmd.Args, " "),
			" with output in file: ", stdout.String(),
			" with err output in file: ", stderr.String(),
			" with status code: ",
			(cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus())

		waitErrChannel <- nil

		close(waitErrChannel)
		close(waitEndChannel)
	}()

	return newlocalTask(cmd, &stdout, &stderr, waitErrChannel, waitEndChannel), nil
}

const killTimeout = 5 * time.Second

// localTask implements Task interface.
type localTask struct {
	waitMutex      sync.Mutex
	cmdHandler     *exec.Cmd
	stdout         *bytes.Buffer
	stderr         *bytes.Buffer
	waitErrChannel chan error
	waitEndChannel chan struct{}
	killTimeout    time.Duration
}

// newlocalTask returns a localTask instance.
func newlocalTask(cmdHandler *exec.Cmd, stdout *bytes.Buffer,
	stderr *bytes.Buffer, waitErrChannel chan error, waitEndChannel chan struct{}) *localTask {
	t := &localTask{
		cmdHandler:     cmdHandler,
		stdout:         stdout,
		stderr:         stderr,
		waitErrChannel: waitErrChannel,
		waitEndChannel: waitEndChannel,
		killTimeout:    killTimeout,
	}
	return t
}

// isTerminated checks if waitEndChannel is closed. If it is closed, it means
// that wait ends, task is in terminated state and ProcessState is not nil.
// ProcessState contains information about an exited process,
// available after call to Wait or Run.
func (task *localTask) isTerminated() bool {
	select {
	case <-task.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	default:
		return false
	}
}

func (task *localTask) getPid() int {
	return task.cmdHandler.Process.Pid
}

func (task *localTask) createStatus() *Status {
	if !task.isTerminated() {
		return nil
	}

	return &Status{
		(task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus(),
		task.stdout.String(),
		task.stderr.String(),
	}
}

func (task *localTask) killTask(sig syscall.Signal) error {
	// We signal the entire process group.
	// The kill syscall interprets a negated PID N as the process group N belongs to.
	log.Debug("Sending ", sig, " to PID ", -task.getPid())
	return syscall.Kill(-task.getPid(), sig)
}

// Stop terminates the local task.
func (task *localTask) Stop() error {
	if task.isTerminated() {
		return nil
	}

	// Sending SIGKILL signal to local task.
	// TODO: Add PID namespace to handle orphan tasks properly.
	err := task.killTask(syscall.SIGKILL)
	if err != nil {
		log.Error(err)
		return err
	}

	// Checking if kill was succesful.
	isTerminated, taskErr := task.Wait(killTimeout)
	if taskErr != nil {
		log.Error(taskErr.Error())
		return taskErr
	}

	if !isTerminated {
		return errors.New("Cannot kill -9 task")
	}

	// No error, task terminated.
	return nil
}

// Status returns a state of the task. If task is terminated it returns the Status as a
// second item in tuple. Otherwise returns nil.
func (task *localTask) Status() (TaskState, *Status) {
	if !task.isTerminated() {
		return RUNNING, nil
	}

	return TERMINATED, task.createStatus()
}

// Wait waits for the command to finish with the given timeout time.
// In case of timeout == 0 there is no timeout for that.
// It returns true if task is terminated.
func (task *localTask) Wait(timeout time.Duration) (bool, error) {
	if task.isTerminated() {
		return true, nil
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case err, ok := <-task.waitErrChannel:
		if err != nil && ok {
			// If channel is not closed and there is error return false and error.
			return false, err
		}

		// If channel is closed or there is no error return true and nil.
		return true, nil
	case <-timeoutChannel:
		// If timeout time exceeded return false and nil.
		return false, nil
	}
}
