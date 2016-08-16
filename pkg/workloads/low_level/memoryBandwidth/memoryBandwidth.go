package memoryBandwidth

import (
	"fmt"
	"path"
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID              = "membw"
	name            = "memBW"
	defaultDuration = 86400 * time.Second
)

// PathFlag represents l3data path flag.
var PathFlag = conf.NewStringFlag(
	"membw_path",
	"Path to Memory Bandwidth binary",
	path.Join(fs.GetSwanWorkloadsPath(), "low-level-aggressors/memBw"),
)

// Config is a struct for MemBw aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultMemBwConfig is a constructor for memBw aggressor Config with default parameters.
func DefaultMemBwConfig() Config {
	return Config{
		Path:     PathFlag.Value(),
		Duration: defaultDuration,
	}
}

// memBw is a launcher for memBw aggressor.
type memBw struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for memBw aggressor.
func New(exec executor.Executor, config Config) executor.Launcher {
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
		return errors.Errorf("launcher configuration is invalid. `duration` value(%d) is lower/equal than/to 0",
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
