package executor

import (
	"github.com/intelsdi-x/swan/pkg/executor/command"
)

// NewLocal returns a Async Executor with Local command instance.
// It's a composition of Async with Local Command.
func NewLocal() Async {
	factory := func() command.Command {
		return command.NewLocal()
	}

	return NewAsync(factory)
}
