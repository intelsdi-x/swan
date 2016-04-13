package executor

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"regexp"
	"strconv"
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
		var stderr bytes.Buffer
		exitCode := -1
		session.Stderr = &stderr
		output, err := session.Output(command)
		if err != nil {
			exitCode, err = getExitCode(err.Error())
			if err != nil {
				log.Error(err.Error())
			}
		}
		statusCh <- Status{exitCode, string(output[:]), stderr.String()}
	}()
	remoteTask := newRemoteTask(session, statusCh)
	return remoteTask, nil
}

func getExitCode(errorMsg string) (int, error) {
	re, err := regexp.Compile(`Process exited with: ([0-9]+).`)
	if err != nil {
		error := fmt.Sprintf(
			"Could not retrieve exit code from output: %s", errorMsg)
		return -1, errors.New(error)
	}
	match := re.FindStringSubmatch(errorMsg)
	if len(match[1]) == 0 {
		error := fmt.Sprintf(
			"Could not retrieve exit code from output: %s", errorMsg)
		return -1, errors.New(error)
	}
	code, _ := strconv.Atoi(match[1])
	return code, nil
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
func (task *remoteTask) Status() (TaskState, *Status) {
	if !task.terminated {
		return RUNNING, nil
	}

	return TERMINATED, &task.status
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
}

// NewRemoteTask returns a RemoteTask instance.
func newRemoteTask(session *ssh.Session, statusCh chan Status) *remoteTask {
	return &remoteTask{
		false,
		session,
		statusCh,
		Status{},
	}
}

// Set task as completed and clean channel
func (task *remoteTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusCh = nil
}
