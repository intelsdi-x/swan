package memcached

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	defaultPort           = 11211
	defaultUser           = "memcached"
	defaultNumThreads     = 4
	defaultMaxMemoryMB    = 64
	defaultNumConnections = 1024
)

// Config is a config for the memcached data caching application v 1.4.25.
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
type Config struct {
	pathToBinary   string
	port           int
	user           string
	numThreads     int
	maxMemoryMB    int
	numConnections int
}

// DefaultMemcachedConfig is a constructor for MemcachedConfig with default parameters.
func DefaultMemcachedConfig(pathToBinary string) Config {
	return Config{
		pathToBinary:   pathToBinary,
		port:           defaultPort,
		user:           defaultUser,
		numThreads:     defaultNumThreads,
		maxMemoryMB:    defaultMaxMemoryMB,
		numConnections: defaultNumConnections,
	}
}

// Memcached is a launcher for the memcached data caching application v 1.4.25.
type Memcached struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for Memcached.
func New(exec executor.Executor, config Config) Memcached {
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
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (m Memcached) Launch() (executor.TaskHandle, error) {
	task, err := m.exec.Execute(m.buildCommand())
	if err != nil {
		return task, err
	}
	// TODO(mpatelcz): we need to assure that memcached is running and
	// operational. This is quick hack to just wait. We need to verify
	// it in more general way (i.e. try to connect to instance).
	time.Sleep(3 * time.Second)
	return task, err
}
