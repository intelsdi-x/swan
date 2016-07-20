package executor

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

// Parallel allows to run same command using same executor multiple times.
// Using Parallel decorator will mix output from all the commands executed.
// You should ALWAYS use this decorator in conjuction with isolation.namespace. See integration tests for use example.
type Parallel struct {
	clones int
}

// NewParallel prepares instance of Executor that allows to run tasks in Parallel.
func NewParallel(clones int) Parallel {
	return Parallel{clones: clones}
}

// Decorate implements isolation.Decorator interface by adding invocation of parallel to a command.
func (p Parallel) Decorate(command string) string {
	var values []interface{}
	values = append(values, p.clones, command)
	decorated := "parallel -j%d sh -c %q --"
	for i := 0; i < p.clones; i++ {
		values = append(values, i)
		decorated += " %d"
	}
	decorated = fmt.Sprintf(decorated, values...)
	logrus.Debug("Running parallelized command: " + decorated)
	logrus.Debug("Be aware that using Parallel decorator will mix output from all the commands executed")

	return decorated
}
