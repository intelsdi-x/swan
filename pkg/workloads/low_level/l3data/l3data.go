package l3data

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/osutil"
	"github.com/intelsdi-x/swan/pkg/swan"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"path"
)

const (
	name            = "L3 Data"
	defaultDuration = 86400 * time.Second
	defaultL3Path   = "low-level-aggressors/l3"
	l3PathEnv       = "SWAN_L3_PATH"
)

// GetPathFromEnvOrDefault returns the l3 binary path from environment variable
// SWAN_L3_PATH or default path in swan directory.
func GetPathFromEnvOrDefault() string {
	return osutil.GetEnvOrDefault(
		l3PathEnv, path.Join(swan.GetSwanWorkloadsPath(), defaultL3Path))
}

// Config is a struct for l3 aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultL3Config is a constructor for l3 aggressor Config with default parameters.
func DefaultL3Config() Config {
	return Config{
		Path:     GetPathFromEnvOrDefault(),
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
