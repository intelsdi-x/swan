package memoryBandwidth

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"path"
)

const (
	name             = "memBW"
	defaultDuration  = 86400 * time.Second
	defaultMemBwPath = "low-level-aggressors/memBw"
	memBwPathEnv     = "SWAN_MEMBW_PATH"
)

// GetPathFromEnvOrDefault fetches the memoryBandwidth binary path from environment variable
// SWAN_MEMBW_PATH or default path in swan directory.
func GetPathFromEnvOrDefault() string {
	return os.GetEnvOrDefault(
		memBwPathEnv, path.Join(fs.GetSwanWorkloadsPath(), defaultMemBwPath))
}

// Config is a struct for MemBw aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultMemBwConfig is a constructor for memBw aggressor Config with default parameters.
func DefaultMemBwConfig() Config {
	return Config{
		Path:     GetPathFromEnvOrDefault(),
		Duration: defaultDuration,
	}
}

// memBw is a launcher for memBw aggressor.
type memBw struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for memBw aggressor.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return memBw{
		exec: exec,
		conf: config,
	}
}

func (m memBw) buildCommand() string {
	return fmt.Sprintf("%s %d", m.conf.Path, int(m.conf.Duration.Seconds()))
}

func (m memBw) verifyConfiguration() error {
	if m.conf.Duration.Seconds() <= 0 {
		return fmt.Errorf("Launcher configuration is invalid. `duration` value(%d) is lower/equal than/to 0",
			int(m.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (m memBw) Launch() (executor.TaskHandle, error) {
	if err := m.verifyConfiguration(); err != nil {
		return nil, err
	}
	return m.exec.Execute(m.buildCommand())
}

// Name returns human readable name for job.
func (m memBw) Name() string {
	return name
}
