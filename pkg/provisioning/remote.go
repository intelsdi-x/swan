package provisioning

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
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

// RemoteTask implements Task interface.
type RemoteTask struct {
	terminated    bool
	session       *ssh.Session
	statusChannel chan Status
	status        Status
}

// NewRemoteTask returns a RemoteTask instance.
func NewRemoteTask(session *ssh.Session, statusChannel chan Status) *RemoteTask {
	return &RemoteTask{
		false,
		session,
		statusChannel,
		Status{},
	}
}

// Set task as completed and clean channel
func (task *RemoteTask) completeTask(status Status) {
	task.terminated = true
	task.status = status
	task.statusChannel = nil
}

// Stop terminates the remote task.
func (task *RemoteTask) Stop() error {
	if task.terminated {
		return errors.New("Task has already completed.")
	}
	err := task.session.Signal("KILL")
	if err != nil {
		return err
	}

	s := <-task.statusChannel
	task.completeTask(s)

	return nil
}

// Status gets status of the remote task.
func (task RemoteTask) Status() Status {
	return task.status
}

// Wait blocks until process is terminated or timeout appeared.
func (task *RemoteTask) Wait(timeoutMs int) bool {
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

// Run runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (remote Remote) Run(command string) (Task, error) {
	statusChannel := make(chan Status)

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
		output, err := session.Output(command)
		if err != nil {
			panic(err)
		}
		//TODO: get exit status
		statusChannel <- Status{0, string(output[:]), ""}
	}()

	remoteTask := NewRemoteTask(session, statusChannel)
	return remoteTask, nil
}
