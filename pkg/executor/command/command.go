package command

import (
	"io"
)

// Command represent common API for ssh.Session and exec.Command.
type Command interface {
	Start(command string, stdout io.Writer, stderr io.Writer) error
	Wait() error
	ExitCode() int
	Kill() error
}
