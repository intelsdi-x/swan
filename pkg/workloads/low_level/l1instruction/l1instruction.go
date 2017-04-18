package l1instruction

import (
	"fmt"
	"math"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID   = "l1i"
	name = "L1 Instruction"

	// {min,max}Intensity are hardcoded values in l1i binary
	minIntensity     = 0
	maxIntensity     = 20
	defaultIntensity = 0 // Most intensive.

	// max int
	maxIterations     = math.MaxInt32
	defaultIterations = maxIterations
)

// PathFlag represents l1i path flag.
var PathFlag = conf.NewStringFlag(
	"l1i_path",
	"Path to L1 instruction binary",
	"l1i",
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
func DefaultL1iConfig() Config {
	return Config{
		Path:       PathFlag.Value(),
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
func New(exec executor.Executor, config Config) executor.Launcher {
	return l1i{
		exec: exec,
		conf: config,
	}
}

func (l l1i) buildCommand() string {
	return fmt.Sprintf("%s %d %d", l.conf.Path, l.conf.Iterations, l.conf.Intensity)
}

func (l l1i) verifyConfiguration() error {
	if l.conf.Intensity > maxIntensity || l.conf.Intensity < minIntensity {
		return errors.Errorf("intensivity value(%d) is out of range <%d;%d>",
			l.conf.Intensity,
			minIntensity,
			maxIntensity)
	}
	if l.conf.Iterations <= 0 || l.conf.Iterations > maxIterations {
		return errors.Errorf("iterations value(%d) should be greater than 0", l.conf.Iterations)
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

// Name returns human readable name for job.
func (l l1i) Name() string {
	return name
}
