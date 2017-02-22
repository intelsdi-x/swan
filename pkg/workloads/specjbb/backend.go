package specjbb

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	name         = "SPECjbb Backend"
	backendJvmID = "specjbbbackend1"
)

var (
	// PathToBinaryForHpFlag specifies path to a SPECjbb2015 jar file for hp job.
	PathToBinaryForHpFlag = conf.NewStringFlag("specjbb_path_hp",
		"Path to SPECjbb jar for high priority job (backend)",
		"/usr/share/specjbb/specjbb2015.jar")
	// PathToPropsFileForHpFlag specifies path to a SPECjbb2015 properties file for hp job.
	PathToPropsFileForHpFlag = conf.NewStringFlag("specjbb_props_path_hp",
		"Path to SPECjbb properties file for high priority job (backend)",
		"/usr/share/specjbb/config/specjbb2015.props")
)

// BackendConfig is a config for a SPECjbb2015 Backend,
type BackendConfig struct {
	JVMOptions
	PathToBinary      string
	ControllerAddress string // ControllerAddress is an address of a SPECjbb controller component ("-Dspecjbb.controller.host=")
	JvmID             string // JvmId is an ID of a JVM dedicated for a Backend (-J <jvmid>)
	WorkerCount       int
}

// DefaultSPECjbbBackendConfig is a constructor for BackendConfig with default parameters.
func DefaultSPECjbbBackendConfig() BackendConfig {
	return BackendConfig{
		JVMOptions:        DefaultJVMOptions(),
		PathToBinary:      PathToBinaryForHpFlag.Value(),
		ControllerAddress: ControllerAddress.Value(),
		JvmID:             backendJvmID,
		WorkerCount:       1,
	}
}

// Backend is a launcher for the SPECjbb2015 Backend.
type Backend struct {
	exec executor.Executor
	conf BackendConfig
}

// NewBackend is a constructor for Backend.
func NewBackend(exec executor.Executor, config BackendConfig) Backend {
	return Backend{
		exec: exec,
		conf: config,
	}
}

// Launch starts the Backend component. It returns a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (b Backend) Launch() (executor.TaskHandle, error) {
	command := getBackendCommand(b.conf)
	task, err := b.exec.Execute(command)
	if err != nil {
		return nil, errors.Wrapf(err, "launch of SPECjbb backend failed. command: %q", command)
	}
	return task, nil
}

// Name returns human readable name for job.
func (b Backend) Name() string {
	return name
}
