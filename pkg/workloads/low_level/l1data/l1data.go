package l1data

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
)

const (
	defaultDuration = 86400 * time.Second
)

// Config is a struct for l1d aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultL1dConfig is a constructor for l1d aggressor Config with default parameters.
func DefaultL1dConfig(pathToBinary string) Config {
	return Config{
		Path:     pathToBinary,
		Duration: defaultDuration,
	}
}

// l1d is a launcher for l1d aggressor.
type l1d struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for l1d aggressor.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return l1d{
		exec: exec,
		conf: config,
	}
}

func (l l1d) buildCommand() string {
	return fmt.Sprintf("%s %d", l.conf.Path, int(l.conf.Duration.Seconds()))
}

func (l l1d) verifyConfiguration() error {
	if l.conf.Duration.Seconds() <= 0 {
		return fmt.Errorf("Launcher configuration is invalid. `duration` value(%d) is lower/equal than/to 0",
			int(l.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (l l1d) Launch() (executor.TaskHandle, error) {
	if err := l.verifyConfiguration(); err != nil {
		return nil, err
	}
	return l.exec.Execute(l.buildCommand())
}
