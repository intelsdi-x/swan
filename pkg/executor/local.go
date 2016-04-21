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

	t := newlocalTask(cmd, &stdout, &stdout)

	return t, err
}

const killTimeout = 5 * time.Second

// localTask implements Task interface.
type localTask struct {
	sync.Mutex
	cmdHandler *exec.Cmd
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer

	killTimeout time.Duration
}

// newlocalTask returns a localTask instance.
func newlocalTask(cmdHandler *exec.Cmd,
	stdout *bytes.Buffer, stderr *bytes.Buffer) *localTask {
	t := &localTask{
		cmdHandler:  cmdHandler,
		stdout:      stdout,
		stderr:      stderr,
		killTimeout: killTimeout,
	}
	return t
}

// isTerminated checks if ProcessState is not nil.
// ProcessState contains information about an exited process,
// available after successful call to Wait or Run.
func (task *localTask) isTerminated() bool {
	return task.cmdHandler.ProcessState != nil
}

func (task *localTask) getPid() int {
	return task.cmdHandler.Process.Pid
}

func (task *localTask) getExitCode() int {
	if !task.isTerminated() {
		panic("Exit code is not available until the task terminated.")
	}

	// If Process exited on his own, show the exitStatus.
	if (task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).Exited() {
		return (task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus()
	}
	// Show what signal caused the termination. (in form of 128 + signal code)
	return 128 + int((task.cmdHandler.ProcessState.Sys().(syscall.WaitStatus)).Signal())
}

func (task *localTask) createStatus() *Status {
	if !task.isTerminated() {
		return nil
	}

	return &Status{
		task.getExitCode(),
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

	// Sending sigterm signal to local task.
	err := task.killTask(syscall.SIGTERM)
	if err != nil {
		log.Error(err)
		return err
	}

	// Wait for task termination with timeout.
	waitErrChannel := make(chan error, 1)
	go func() {
		waitErrChannel <- task.Wait()
	}()

	select {
	case taskErr := <-waitErrChannel:
		if taskErr != nil {
			log.Error(taskErr.Error())
			return taskErr
		}
	case <-time.After(task.killTimeout):
		// We need to send SIGKILL to the entire process group, since SIGTERM was not enough.
		err := task.killTask(syscall.SIGKILL)
		if err != nil {
			log.Error(err)
			return err
		}

		// Wait again with the timeout.
		select {
		case taskErr := <-waitErrChannel:
			if taskErr != nil {
				log.Error(taskErr.Error())
				return taskErr
			}
		case <-time.After(task.killTimeout):
			return errors.New("Cannot kill -9 task")
		}
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

// Wait blocks until process is terminated.
func (task *localTask) Wait() error {
	// This function needs to be performed only by one go routine at once.
	task.Lock()
	defer task.Unlock()

	if task.isTerminated() {
		return nil
	}

	// Wait for task completion.
	// NOTE: Wait() returns an error. We grab the process state in any case
	// (success or failure) below, so the error object matters less in the
	// status handling for now.
	if err := task.cmdHandler.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			// In case of NON Exit Errors we are sure that task does not terminate so return error.
			return err
		}
	}

	log.Debug(
		"Ended ", strings.Join(task.cmdHandler.Args, " "),
		" with output in file: ", task.stdout.String(),
		" with err output in file: ", task.stderr.String(),
		" with status code: ", task.getExitCode())

	return nil
}
