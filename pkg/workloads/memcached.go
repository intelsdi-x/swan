package workloads

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	memcachedDefaultPort           = 11211
	memcachedDefaultUser           = "memcached"
	memcachedDefaultNumThreads     = 4
	memcachedDefaultMaxMemoryMB    = 64
	memcachedDefaultNumConnections = 1024
)

// MemcachedConfig is a config for the memcached data caching application v 1.4.25.
// memcached 1.4.25, supported options:
// -p <num>      TCP port number to listen on (default: 11211)
// -s <file>     UNIX socket path to listen on (disables network support)
// -u <username> assume identity of <username> (only when run as root)
// -m <num>      max memory to use for items in megabytes (default: 64 MB)
// -c <num>      max simultaneous connections (default: 1024)
//-v            verbose (print errors/warnings while in event loop)
//-vv           very verbose (also print client commands/reponses)
//-vvv          extremely verbose (also print internal state transitions)
//-t <num>      number of threads to use (default: 4)
type MemcachedConfig struct {
	pathToBinary   string
	port           int
	user           string
	numThreads     int
	maxMemoryMB    int
	numConnections int
}

// DefaultMemcachedConfig is a constructor for MemcachedConfig with default parameters.
func DefaultMemcachedConfig(pathToBinary string) MemcachedConfig {
	return MemcachedConfig{
		pathToBinary,
		memcachedDefaultPort,
		memcachedDefaultUser,
		memcachedDefaultNumThreads,
		memcachedDefaultMaxMemoryMB,
		memcachedDefaultNumConnections,
	}
}

// Memcached is a launcher for the memcached data caching application v 1.4.25.
type Memcached struct {
	exec executor.Executor
	conf MemcachedConfig
}

// NewMemcached is a constructor for Memcached.
func NewMemcached(exec executor.Executor, config MemcachedConfig) Memcached {
	return Memcached{
		exec: exec,
		conf: config,
	}

}
func (m Memcached) buildCommand() string {
	return fmt.Sprint(m.conf.pathToBinary,
		" -p ", m.conf.port,
		" -u ", m.conf.user,
		" -t ", m.conf.numThreads,
		" -m ", m.conf.maxMemoryMB,
		" -c ", m.conf.numConnections)
}

// Launch starts the workload (process or group of processes). It returns a workload
// represented as a Task instance.
// Error is returned when Launcher is unable to start a job.
func (m Memcached) Launch() (executor.Task, error) {
	return m.exec.Execute(m.buildCommand())
}
