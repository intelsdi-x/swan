package executor

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Local provisioning is responsible for providing the execution environment
// on local machine via exec.Command.
// It runs command as current user.
type Local struct {
	OutputPrefix string
}

// NewLocal returns a Local instance.
func NewLocal() Local {
	return Local{OutputPrefix: "unique"}
}

// Execute runs the command given as input.
// Returned Task is able to stop & monitor the provisioned process.
func (l Local) Execute(command string) (Task, error) {
	log.Debug("Starting locally ", command)

	cmd := exec.Command("sh", "-c", command)
	// It is important to set additional Process Group ID for parent process and his children
	// to have ability to kill all the children processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var (
		stdoutFile, stderrFile *os.File
		err                    error
	)
	// Create temporary output files.
	if l.OutputPrefix != "unique" {
		stdoutFile, err = os.OpenFile("./"+l.OutputPrefix+"_stdout",
			os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, err
		}
		stderrFile, err = os.OpenFile("./"+l.OutputPrefix+"_stderr",
			os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		stdoutFile, err = ioutil.TempFile("./", "swan_local_executor_stdout_")
		if err != nil {
			return nil, err
		}
		stderrFile, err = ioutil.TempFile("./", "swan_local_executor_stderr_")
		if err != nil {
			return nil, err
		}
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

	return newlocalTask(cmd, stdoutFile, stderrFile, waitEndChannel), nil
}

// localTask implements Task interface.
type localTask struct {
	cmdHandler     *exec.Cmd
	stdoutFile     *os.File
	stderrFile     *os.File
	waitEndChannel chan struct{}
}

// newlocalTask returns a localTask instance.
func newlocalTask(cmdHandler *exec.Cmd, stdoutFile *os.File, stderrFile *os.File,
	waitEndChannel chan struct{}) *localTask {
	t := &localTask{
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
		task.stdoutFile.Name(),
		task.stderrFile.Name(),
	}
}

// Stop terminates the local task.
func (task *localTask) Stop() error {
	if task.isTerminated() {
		return nil
	}

	// Sending SIGKILL signal to local task.
	// TODO: Add PID namespace to handle orphan tasks properly.
	log.Debug("Sending ", syscall.SIGKILL, " to PID ", -task.getPid())
	err := syscall.Kill(-task.getPid(), syscall.SIGKILL)
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

// Status returns a state of the task. If task is terminated it returns the Status as a
// second item in tuple. Otherwise returns nil.
func (task *localTask) Status() (TaskState, *Status) {
	if !task.isTerminated() {
		return RUNNING, nil
	}

	return TERMINATED, task.createStatus()
}

// Stdout returns io.Reader to stdout file.
func (task *localTask) Stdout() (io.Reader, error) {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return nil, err
	}

	task.stdoutFile.Seek(0, os.SEEK_SET)
	return task.stdoutFile, nil
}

// Stderr returns io.Reader to stderr file.
func (task *localTask) Stderr() (io.Reader, error) {
	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return nil, err
	}

	task.stderrFile.Seek(0, os.SEEK_SET)
	return task.stderrFile, nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *localTask) Clean() error {
	// Close stdout.
	stdoutErr := task.stdoutFile.Close()

	// Close stderr.
	stderrErr := task.stderrFile.Close()

	if stdoutErr != nil {
		return stdoutErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (task *localTask) EraseOutput() error {
	// Remove stdout file.
	if err := os.Remove(task.stdoutFile.Name()); err != nil {
		return err
	}

	// Remove stderr file.
	if err := os.Remove(task.stderrFile.Name()); err != nil {
		return err
	}

	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (task *localTask) Wait(timeout time.Duration) bool {
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
