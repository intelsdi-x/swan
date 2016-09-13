package loadgenerator

import (
	"github.com/intelsdi-x/athena/pkg/executor"
)

const (
	defaultJVMNId = 1
)

// Config is a config for a SPECjbb2015 Transaction Injector,
// Supported options:
// IP - property "-Dspecjbb.controller.host=" - IP address of a SPECjbb controller component (default:127.0.0.1)
// JvmId - argument -J JVM<num> - id of a JVM dedicated for a transaction injector
type Config struct {
	ControllerIP string
	JvmID        int
}

// DefaultSPECjbbTxIConfig is a constructor for Config with default parameters.
func DefaultSPECjbbTxIConfig() Config {
	return Config{
		ControllerIP: IPFlag.Value(),
		JvmID:        defaultJVMNId,
	}
}

// TxI is a launcher for the SPECjbb2015 Transaction Injector.
type TxI struct {
	Exec executor.Executor
	Conf Config
}

// NewTxI is a constructor for Txl.
func NewTxI(exec executor.Executor, config Config) TxI {
	return TxI{
		Exec: exec,
		Conf: config,
	}
}
