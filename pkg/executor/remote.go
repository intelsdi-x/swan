package executor

import (
	"github.com/intelsdi-x/swan/pkg/executor/command"
)

// NewRemote returns a Async Executor with Remote command instance.
// It's a composition of Async with Remote Command.
func NewRemote(sshConfig command.SSHConfig) Async {
	factory := func() command.Command {
		return command.NewRemote(sshConfig)
	}
	return NewAsync(factory)
}
