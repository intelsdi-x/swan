package executor

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig SSHConfig
}

// NewRemote returns a Local instance.
func NewRemote(sshConfig SSHConfig) *Remote {
	return &Remote{
		sshConfig,
	}
}

// Execute runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (remote Remote) Execute(command string) (Task, error) {
	statusCh := make(chan Status)
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
	connection, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", remote.sshConfig.host, remote.sshConfig.port),
		remote.sshConfig.clientConfig,
	)
	if err != nil {
		return nil, err
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, err
	}

	terminal := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 40, terminal); err != nil {
		session.Close()
		return nil, err
	}

	go func() {
		exitCode := 0

		session.Stderr = stderrFile
		session.Stdout = stdoutFile
		err := session.Run(command)

		if err != nil {
			exitCode = err.(*ssh.ExitError).Waitmsg.ExitStatus()
		}
		statusCh <- Status{&exitCode}
	}()

	remoteTask := newRemoteTask(session, statusCh, stdoutFile, stderrFile)
	return remoteTask, nil
}

// RemoteTask implements Task interface.
type remoteTask struct {
	terminated bool
	session    *ssh.Session
	statusCh   chan Status
	status     Status
	stdoutFile *os.File
	stderrFile *os.File
}

// NewRemoteTask returns a RemoteTask instance.
func newRemoteTask(session *ssh.Session, statusCh chan Status, stdoutFile *os.File, stderrFile *os.File) *remoteTask {
	return &remoteTask{
		false,
		session,
		statusCh,
		Status{},
		stdoutFile,
		stderrFile,
	}
}

// Stdout returns io.Reader to stdout file.
func (task *remoteTask) Stdout() io.Reader {
	r := io.Reader(task.stdoutFile)
	return r
}

// Stderr returns io.Reader to stderr file.
func (task *remoteTask) Stderr() io.Reader {
	r := io.Reader(task.stderrFile)
	return r
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *remoteTask) Clean() error {
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

// Stop terminates the remote task.
func (task *remoteTask) Stop() error {
	if task.terminated {
		return nil
	}
	err := task.session.Signal(ssh.SIGKILL)
	if err != nil {
		return err
	}

	s := <-task.statusCh
	task.completeTask(s)

	return nil
}

// Status gets task state and status of the remote task.
func (task *remoteTask) Status() (TaskState, *int) {
	if !task.terminated {
		return RUNNING, nil
	}

	return TERMINATED, task.status.ExitCode
}

// Wait blocks until process is terminated or timeout appeared.
// Returns true when process terminates before timeout, otherwise false.
func (task *remoteTask) Wait(timeoutMs int) bool {
	if task.terminated {
		return true
	}

	if timeoutMs == 0 {
		s := <-task.statusCh
		task.completeTask(s)
		return true
	}

	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	result := true
	select {
	case s := <-task.statusCh:
		task.completeTask(s)
	case <-time.After(timeoutDuration):
		result = false
	}

	return result
}

// Set task as completed and clean channel
func (task *remoteTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusCh = nil
}
