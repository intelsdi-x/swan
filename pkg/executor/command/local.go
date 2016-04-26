package command

import (
	log "github.com/Sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
)

// Local command is responsible for providing the execution environment
// on local machine via exec.Command.
type Local struct {
	cmd *exec.Cmd
}

// NewLocal returns a Local instance.
func NewLocal() *Local {
	return &Local{}
}

// Start starts the command.
func (l *Local) Start(command string, stdout io.Writer, stderr io.Writer) error {
	l.cmd = exec.Command("sh", "-c", command)
	// It is important to set additional Process Group ID for parent process and his children
	// to have ability to kill all the children processes.
	l.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	l.cmd.Stdout = stdout
	l.cmd.Stderr = stderr

	return l.cmd.Start()
}

// Wait waits synchronously for task.
func (l *Local) Wait() error {
	// Wait for task completion.
	// NOTE: Wait() returns an error. We grab the process state in any case
	// (success or failure) below, so the error object matters less in the
	// status handling for now.
	if err := l.cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			// In case of NON Exit Errors we are not sure if task does
			// terminate so panic.
			return err
		}
	}
	return nil
}

func (l *Local) getPid() int {
	return l.cmd.Process.Pid
}

// ExitCode returns ExitCode. It is only allowed to use this method when
// task is terminated.
func (l *Local) ExitCode() int {
	return (l.cmd.ProcessState.Sys().(syscall.WaitStatus)).ExitStatus()
}

// Kill sends SIGKILL signal to task.
func (l *Local) Kill() error {
	// We signal the entire process group.
	// The kill syscall interprets a negated PID N as the process group N belongs to.
	log.Debug("Sending ", syscall.SIGKILL, " to PID ", -l.getPid())
	return syscall.Kill(-l.getPid(), syscall.SIGKILL)
}
