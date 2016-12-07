package specjbb

import (
	"fmt"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/pkg/errors"
)

const (
	name          = "SPECjbb Backend"
	defaultJVMNId = 1
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
// Supported options:
// IP - property "-Dspecjbb.controller.host=" - IP address of a SPECjbb controller component (default:127.0.0.1)
// JvmId - argument -J JVM<num> - id of a JVM dedicated for a Backend
type BackendConfig struct {
	PathToBinary string
	IP           string
	JvmID        int
}

// DefaultSPECjbbBackendConfig is a constructor for Config with default parameters.
func DefaultSPECjbbBackendConfig() BackendConfig {
	return BackendConfig{
		PathToBinary: PathToBinaryForHpFlag.Value(),
		IP:           IPFlag.Value(),
		JvmID:        TxICountFlag.Value() + defaultJVMNId, // Backend JVM Id is always one more than number of TxI components.
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

func (b Backend) buildCommand() string {
	return fmt.Sprint("java -jar",
		" -Dcom.sun.management.jmxremote.port=5555",
		" -Dcom.sun.management.jmxremote.ssl=false",
		" -Dcom.sun.management.jmxremote.authenticate=false",
		" -Djava.net.preferIPv4Stack=true",
		" -Xms5g -Xmx5g",                    // allocate whole heap available; docs: For best performance, set -Xms to the same size as the maximum heap size
		" -XX:NativeMemoryTracking=summary", // memory monitoring purposes
		" -server",                          // compilation takes more time but offers additional optimizations
		ControllerHostProperty, b.conf.IP,
		" ", b.conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J JVM", b.conf.JvmID,
		" -p ", PathToPropsFileForHpFlag.Value(),
	)
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
