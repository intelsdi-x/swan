package provisioning

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
	"bytes"
	"regexp"
	"strconv"
)

// RemoteTask implements Task interface.
type RemoteTask struct{
	isTerminated bool
	session *ssh.Session
	statusCh chan Status
	status Status
}

// NewRemoteTask returns a RemoteTask instance.
func NewRemoteTask(session *ssh.Session, statusCh chan Status) *RemoteTask {
    return &RemoteTask{
	    false,
	    session,
	    statusCh,
	    Status{},
    }
}

// Set task as completed and clean channel
func (task *RemoteTask) completeTask(status Status) {
	task.isTerminated = true
	task.status = status
	task.statusCh = nil
}

// Stop terminates the remote task.
func (task *RemoteTask) Stop() error {
    if task.isTerminated {
	    return errors.New("Task has already completed.")
    }
    err := task.session.Signal("KILL");
	if err != nil {
		return err
	}

	s := <-task.statusCh
	task.completeTask(s)

	return nil
}

// Status gets status of the remote task.
func (task RemoteTask) Status() Status {
	return task.status
}

// Wait blocks until process is terminated or timeout appeared.
func (task *RemoteTask) Wait(timeoutMs int) bool {
	if (task.isTerminated) {
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

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct{
	sshConfig SshConfig
}

// NewRemote returns a Local instance.
func NewRemote(sshConfig SshConfig) *Remote {
	return &Remote{
		sshConfig,
	}
}

// getExitCode gets exit code from the error message given as input.
func getExitCode(errorMsg string) int{
	re := regexp.MustCompile(`Process exited with: ([0-9]+).`)
	match := re.FindStringSubmatch(errorMsg)
	code, _ := strconv.Atoi(match[0])
	return code
}

// Run runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (remote Remote) Run(command string) (Task, error) {
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
		statusChannel <- Status{statusCode, string(output[:]), stderr.String()}
	}()
	remoteTask := NewRemoteTask(session, statusCh)
	return remoteTask, nil
}
