package executor

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
	"bytes"
	"regexp"
	"strconv"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct{
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

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", remote.sshConfig.host, remote.sshConfig.port),
		remote.sshConfig.clientConfig)
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
		statusCode := 0
		session.Stderr = &stderr
		output, err := session.Output(command)
		if err != nil {
			statusCode = getExitCode(err.Error())
		}
		statusCh <- Status{statusCode, string(output[:]), stderr.String()}
	}()
	remoteTask := newRemoteTask(session, statusCh)
	return remoteTask, nil
}

// getExitCode gets exit code from the error message given as input.
func getExitCode(errorMsg string) int{
	re := regexp.MustCompile(`Process exited with: ([0-9]+).`)
	match := re.FindStringSubmatch(errorMsg)
	if len(match[1]) == 0{
		panic(errors.New("Exit code not found"))
	}
	code, _ := strconv.Atoi(match[1])
	return code
}

// Stop terminates the remote task.
func (task *remoteTask) Stop() error {
	if task.terminated {
		return errors.New("Task has already completed.")
	}
	err := task.session.Signal(ssh.SIGKILL);
	if err != nil {
		return err
	}

	s := <-task.statusCh
	task.completeTask(s)

	return nil
}

// Status gets status of the remote task.
func (task *remoteTask) Status() (TaskState, *Status) {
	if !task.terminated {
		return RUNNING, nil
	}

	return TERMINATED, &task.status
}

// Wait blocks until process is terminated or timeout appeared.
func (task *remoteTask) Wait(timeoutMs int) bool {
	if (task.terminated) {
		return false
	}

	if (timeoutMs == 0) {
		s := <-task.statusCh
		task.completeTask(s)

	} else {
		timeoutDuration := time.Duration(timeoutMs) * time.Millisecond

		select {
		case s := <-task.statusCh:
			task.completeTask(s)
		case <-time.After(timeoutDuration):
			return true
		}
	}

	return false
}

// RemoteTask implements Task interface.
type remoteTask struct{
	terminated bool
	session *ssh.Session
	statusCh chan Status
	status Status
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
