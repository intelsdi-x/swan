package stressng

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID = "stressng"
)

// l1d is a launcher for stress-ng aggressor.
type stressng struct {
	executor  executor.Executor
	arguments string
}

// New is a constructor for stress-ng aggressor.
func New(executor executor.Executor, arguments string) executor.Launcher {
	return stressng{
		executor:  executor,
		arguments: arguments,
	}
}

// Launch starts a workload.
func (s stressng) Launch() (executor.TaskHandle, error) {
	return s.executor.Execute(fmt.Sprintf("stress-ng %s", s.arguments))
}

// Name returns readable name.
func (s stressng) Name() string {
	return "Stress-ng"
}
