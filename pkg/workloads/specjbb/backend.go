package specjbb

import (
	"fmt"
	"path"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/pkg/errors"
)

const (
	name          = "SPECjbb Backend"
	defaultJVMNId = 1
)

var (
	// PathToBinaryFlagHp specifies path to a SPECjbb2015 jar file for hp job.
	PathToBinaryFlagHp = conf.NewStringFlag("specjbb_path_hp", "Path to SPECjbb jar for high priority job (backend)",
		path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "specjbb2015.jar"))
	// PathToPropsFileFlagHp specifies path to a SPECjbb2015 properties file for hp job.
	PathToPropsFileFlagHp = conf.NewStringFlag("specjbb_props_path_hp", "Path to SPECjbb properties file for high priority job (backend)",
		path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "config", "specjbb2015.props"))
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
		PathToBinary: PathToBinaryFlagHp.Value(),
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
		ControllerHostProperty, b.conf.IP,
		" ", b.conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J JVM", b.conf.JvmID,
		" -p ", PathToPropsFileFlagHp.Value())
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
