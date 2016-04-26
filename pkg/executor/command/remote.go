package command

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
)

// Remote command is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig SSHConfig
	session   *ssh.Session
	exitCode  int
}

// NewRemote returns a Local instance.
func NewRemote(sshConfig SSHConfig) *Remote {
	return &Remote{
		sshConfig: sshConfig,
	}
}

// Start runs the command given as input.
func (r *Remote) Start(command string, stdout io.Writer, stderr io.Writer) error {
	connection, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", r.sshConfig.host, r.sshConfig.port),
		r.sshConfig.clientConfig,
	)
	if err != nil {
		return err
	}

	r.session, err = connection.NewSession()
	if err != nil {
		return err
	}

	terminal := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := r.session.RequestPty("xterm", 80, 40, terminal); err != nil {
		r.session.Close()
		return err
	}

	r.session.Stdout = stdout
	r.session.Stderr = stderr

	return r.session.Start(command)
}

// Wait waits synchronously for task.
func (r *Remote) Wait() error {
	// Wait for task completion.
	err := r.session.Wait()

	waitMsg, ok := err.(*ssh.ExitError)
	if !ok {
		return err
	}

	r.exitCode = waitMsg.Waitmsg.ExitStatus()
	return nil
}

// ExitCode returns ExitCode. It is only allowed to use this method when
// task is terminated.
func (r *Remote) ExitCode() int {
	return r.exitCode
}

// Kill sends SIGKILL signal to task.
func (r *Remote) Kill() error {
	return r.session.Signal(ssh.SIGKILL)
}
