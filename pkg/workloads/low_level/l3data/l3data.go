package l3data

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"path"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID              = "l3d"
	name            = "L3 Data"
	defaultDuration = 86400 * time.Second
)

// PathFlag represents l3data path flag.
var PathFlag = conf.NewStringFlag(
	"l3_path",
	"Path to L3 Data binary",
	path.Join(fs.GetSwanWorkloadsPath(), "low-level-aggressors/l3"),
)

// Config is a struct for l3 aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultL3Config is a constructor for l3 aggressor Config with default parameters.
func DefaultL3Config() Config {
	return Config{
		Path:     PathFlag.Value(),
		Duration: defaultDuration,
	}
}

// l3 is a launcher for l3 aggressor.
type l3 struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for l3 aggressor.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return l3{
		exec: exec,
		conf: config,
	}
}

func (l l3) buildCommand() string {
	return fmt.Sprintf("%s %d", l.conf.Path, int(l.conf.Duration.Seconds()))
}

func (l l3) verifyConfiguration() error {
	if l.conf.Duration.Seconds() <= 0 {
		return fmt.Errorf("Launcher configuration is invalid. `duration` value(%v) is lower/equal than/to 0",
			int(l.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (l l3) Launch() (executor.TaskHandle, error) {
	if err := l.verifyConfiguration(); err != nil {
		return nil, err
	}
	return l.exec.Execute(l.buildCommand())
}

// Name returns human readable name for job.
func (l l3) Name() string {
	return name
}
