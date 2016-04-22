package executor

import (
	"fmt"
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
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	stdoutFile, err := ioutil.TempFile(pwd, "stdout")
	if err != nil {
		return nil, err
	}
	stderrFile, err := ioutil.TempFile(pwd, "stderr")
	if err != nil {
		return nil, err
	}

	log.Debug("Starting ", command)

	cmd := exec.Command("sh", "-c", command)
	fmt.Print(stdoutFile.Name())
	// It is important to set additional Process Group ID for parent process and his children
	// to have ability to kill all the children processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

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
			" with output in file: ", stdoutFile.Name(),
			" with err output in file: ", stderrFile.Name(),
			" with status code: ", exitCode)

		statusChannel <- Status{
			&exitCode,
		}
	}()

	taskPid := taskPID(cmd.Process.Pid)

	t := newlocalTask(taskPid, statusChannel, stdoutFile, stderrFile)

	return t, err
}

// localTask implements Task interface.
type localTask struct {
	pid           taskPID
	statusChannel chan Status
	status        Status
	terminated    bool
	stdoutFile    *os.File
	stderrFile    *os.File
}

// newlocalTask returns a localTask instance.
func newlocalTask(pid taskPID, statusChannel chan Status, stdoutFile *os.File, stderrFile *os.File) *localTask {
	t := &localTask{
		pid,
		statusChannel,
		Status{},
		false,
		stdoutFile,
		stderrFile,
	}
	return t
}

// Stdout returns io.Reader to stdout file.
func (task *localTask) Stdout() io.Reader {
	fmt.Print("X", task.stdoutFile.Name())
	r := io.Reader(task.stdoutFile)
	return r
}

// Stderr returns io.Reader to stderr file.
func (task *localTask) Stderr() io.Reader {
	r := io.Reader(task.stderrFile)
	return r
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *localTask) Clean() error {
	//TODO: fix errors returned
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
