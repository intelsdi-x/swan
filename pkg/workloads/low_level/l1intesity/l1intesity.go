package l1intesity

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
)

const (
	defaultIterations = 10
	defaultIntensity  = 20
	// {min,max}Intensity are hardcoded values in l1i binary
	// For further information look inside l1i.c which can be found in github.com/intelsdi-x/swan repository
	minIntensity = 1
	maxIntensity = 20
)

// Config is a struct for l1i aggressor configuration.
type Config struct {
	Path string
	// Intensity means level(in range <1;20>) of L1 load.
	Intensity int
	// Iteration means how many L1 load should be executed.
	Iterations int
}

// DefaultL1iConfig is a constructor for l1i aggressor Config with default parameters.
func DefaultL1iConfig(pathToBinary string) Config {
	return Config{
		Path:       pathToBinary,
		Intensity:  defaultIntensity,
		Iterations: defaultIterations,
	}
}

// l1i is a launcher for l1i aggressor.
type l1i struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for l1i aggressor.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return l1i{
		exec: exec,
		conf: config,
	}
}

func (l l1i) buildCommand() string {
	return fmt.Sprintf("%s %v %v", l.conf.Path, l.conf.Iterations, l.conf.Intensity)
}

func (l l1i) verifyConfiguration() error {
	if l.conf.Intensity > maxIntensity || l.conf.Intensity < minIntensity {
		return fmt.Errorf("Intensivity value(%d) is out of range <%d;%d>",
			l.conf.Intensity,
			minIntensity,
			maxIntensity)
	}
	if l.conf.Iterations <= 0 {
		return fmt.Errorf("Iterations value(%d) should be greater than 0", l.conf.Iterations)
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (l l1i) Launch() (executor.TaskHandle, error) {
	if err := l.verifyConfiguration(); err != nil {
		return nil, err
	}
	return l.exec.Execute(l.buildCommand())
}
