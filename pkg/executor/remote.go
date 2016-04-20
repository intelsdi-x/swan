package executor

import (
	"bytes"
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
	stdoutDir, err := ioutil.TempFile("/tmp/", "stdout")
	if err != nil {
		return nil, err
	}
	stderrDir, err := ioutil.TempFile("/tmp/", "stderr")
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
		var stderr bytes.Buffer

		session.Stderr = &stderr
		output, err := session.Output(command)

		ioutil.WriteFile(stdoutDir.Name(), output, 600)
		ioutil.WriteFile(stderrDir.Name(), stderr.Bytes(), 600)

		if err != nil {
			exitCode = err.(*ssh.ExitError).Waitmsg.ExitStatus()
		}
		statusCh <- Status{&exitCode}
	}()
	remoteTask := newRemoteTask(session, statusCh, stdoutDir.Name(), stderrDir.Name())
	return remoteTask, nil
}

// Stdout returns io.Reader to stdout file.
func (task *remoteTask) Stdout() (io.Reader, error) {
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
func (task *remoteTask) Stderr() (io.Reader, error) {
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
func (task *remoteTask) Clean() error {
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

func (task *remoteTask) GetStdoutDir() (string, error) {
	if _, err := os.Stat(task.stdoutDir); err != nil {
		return "", err
	}
	return task.stdoutDir, nil
}

func (task *remoteTask) GetStderrDir() (string, error) {
	if _, err := os.Stat(task.stderrDir); err != nil {
		return "", err
	}
	return task.stderrDir, nil
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

// RemoteTask implements Task interface.
type remoteTask struct {
	terminated bool
	session    *ssh.Session
	statusCh   chan Status
	status     Status
	stdoutDir  string
	stderrDir  string
}

// NewRemoteTask returns a RemoteTask instance.
func newRemoteTask(session *ssh.Session, statusCh chan Status, stdoutDir string, stderrDir string) *remoteTask {
	return &remoteTask{
		false,
		session,
		statusCh,
		Status{},
		stdoutDir,
		stderrDir,
	}
}

// Set task as completed and clean channel
func (task *remoteTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusCh = nil
}
