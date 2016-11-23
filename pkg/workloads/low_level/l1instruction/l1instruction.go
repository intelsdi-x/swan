package l1instruction

import (
	"fmt"
	"path"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID               = "l1i"
	name             = "L1 Instruction"
	defaultIntensity = 19

	// {min,max}Intensity are hardcoded values in l1i binary
	// For further information look inside l1i.c which can be found in github.com/intelsdi-x/swan
	// repository.
	minIntensity = 0
	maxIntensity = 20

	// -1 (or absence of iteration argument) means infinite iterations in l1i.
	infiniteIterations = -1
	defaultIterations  = infiniteIterations
)

// PathFlag represents l1i path flag.
var PathFlag = conf.NewStringFlag(
	"l1i_path",
	"Path to L1 instruction binary",
	path.Join(fs.GetSwanWorkloadsPath(), "low-level-aggressors/l1i"),
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
	return fmt.Sprintf("%s %d %d", l.conf.Path, l.conf.Intensity, l.conf.Iterations)
}

func (l l1i) verifyConfiguration() error {
	if l.conf.Intensity > maxIntensity || l.conf.Intensity < minIntensity {
		return errors.Errorf("intensivity value(%d) is out of range <%d;%d>",
			l.conf.Intensity,
			minIntensity,
			maxIntensity)
	}
	if l.conf.Iterations <= 0 && l.conf.Iterations != infiniteIterations {
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
