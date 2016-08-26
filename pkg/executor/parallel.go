package executor

import (
	"fmt"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/isolation"
)

// Parallel allows to run same command using same executor multiple times.
// Using Parallel decorator will mix output from all the commands executed.
// Parallel is run in new PID namespace (using isolation.Namespace) as children might not be killed correctly otherwise.
type Parallel struct {
	numberOfClones int
}

// NewParallel prepares instance of Decorator that allows to ran tasks in parallel.
func NewParallel(numberOfClones int) Parallel {
	return Parallel{numberOfClones: numberOfClones}
}

// Decorate implements isolation.Decorator interface by adding invocation of parallel to a command.
func (p Parallel) Decorate(command string) string {
	logrus.Debugf("Attempting to run command %q %d times", command, p.numberOfClones)
	var values []interface{}
	values = append(values, p.numberOfClones, command)
	decorated := "parallel -j%d sh -c %q --"
	for i := 0; i < p.numberOfClones; i++ {
		values = append(values, i)
		decorated += " %d"
	}
	// You need to run parallel in new PID namespace to make sure that all the children are killed.
	unshare, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
	if err != nil {
		logrus.Errorf("Impossible to create namespace decorator: %q", err)
		return command
	}
	decorated = unshare.Decorate(fmt.Sprintf(decorated, values...))
	logrus.Debugf("Parallelized command prepared: %q", decorated)
	logrus.Debug("Be aware that using Parallel decorator will mix output from all the commands executed")

	return decorated
}
