package backend

import (
	"fmt"

	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/loadgenerator"
	"github.com/pkg/errors"
)

const (
	name          = "SPECjbb Backend"
	defaultJVMNId = 1
)

// Config is a config for a SPECjbb2015 Backend,
// Supported options:
// IP - property "-Dspecjbb.controller.host=" - IP address of a SPECjbb controller component (default:127.0.0.1)
// JvmId - argument -J JVM<num> - id of a JVM dedicated for a Backend
type Config struct {
	PathToBinary string
	IP           string
	JvmID        int
}

// DefaultSPECjbbBackendConfig is a constructor for Config with default parameters.
func DefaultSPECjbbBackendConfig() Config {
	return Config{
		PathToBinary: loadgenerator.PathToBinaryFlag.Value(),
		IP:           loadgenerator.IPFlag.Value(),
		JvmID:        loadgenerator.TxICountFlag.Value() + defaultJVMNId, // Backend JVM Id is always one more than number of TxI components.
	}
}

// Backend is a launcher for the SPECjbb2015 Backend.
type Backend struct {
	exec executor.Executor
	conf Config
}

// NewBackend is a constructor for Backend.
func NewBackend(exec executor.Executor, config Config) Backend {
	return Backend{
		exec: exec,
		conf: config,
	}
}

func (b Backend) buildCommand() string {
	return fmt.Sprint("java -jar",
		loadgenerator.ControllerHostProperty, b.conf.IP,
		" ", b.conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J JVM", b.conf.JvmID,
		" -p ", loadgenerator.PathToPropsFileFlag.Value())
}

// Launch starts the Backend component. It returns a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (b Backend) Launch() (executor.TaskHandle, error) {
	task, err := b.exec.Execute(b.buildCommand())
	if err != nil {
		return nil, errors.Wrapf(err, "launch of SPECjbb backend failed. command: %q", b.buildCommand())
	}
	return task, nil
}

// Name returns human readable name for job.
func (b Backend) Name() string {
	return name
}
