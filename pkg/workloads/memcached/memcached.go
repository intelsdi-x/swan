package memcached

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/intelsdi-x/athena/pkg/utils/netutil"
)

const (
	name = "Memcached"
	// DefaultPort represents default memcached port.
	defaultPort           = 11211
	defaultUser           = "memcached"
	defaultNumThreads     = 4
	defaultMaxMemoryMB    = 64
	defaultNumConnections = 1024
	defaultListenIP       = "127.0.0.1"
)

var (
	pathFlag = conf.NewStringFlag("memcached_path", "Path to memcached binary",
		path.Join(fs.GetSwanWorkloadsPath(), "data_caching/memcached/memcached-1.4.25/build/memcached"))
	// PortFlag returns port which will be specified for workload services as endpoints.
	PortFlag = conf.NewIntFlag("memcached_port", "Port for memcached to listen on. (-p)", defaultPort)
	// IPFlag returns IP which will be specified for workload services as endpoints.
	IPFlag             = conf.NewIPFlag("memcached_ip", "IP of interface memcached is listening on.", defaultListenIP)
	userFlag           = conf.NewStringFlag("memcached_user", "Username for memcached process (-u)", defaultUser)
	numThreadsFlag     = conf.NewIntFlag("memcached_threads", "Number of threads for mutilate (-t)", defaultNumThreads)
	maxConnectionsFlag = conf.NewIntFlag("memcached_connections", "Number of maximum connections for mutilate (-c)", defaultNumConnections)
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
	PathToBinary   string
	Port           int
	User           string
	NumThreads     int
	MaxMemoryMB    int
	NumConnections int
	IP             string
}

// DefaultMemcachedConfig is a constructor for MemcachedConfig with default parameters.
func DefaultMemcachedConfig() Config {
	return Config{
		PathToBinary:   pathFlag.Value(),
		Port:           PortFlag.Value(),
		User:           userFlag.Value(),
		NumThreads:     numThreadsFlag.Value(),
		MaxMemoryMB:    defaultMaxMemoryMB,
		NumConnections: maxConnectionsFlag.Value(),
		IP:             IPFlag.Value(),
	}
}

// Memcached is a launcher for the memcached data caching application v 1.4.25.
type Memcached struct {
	exec          executor.Executor
	conf          Config
	isMemcachedUp netutil.IsListeningFunction // For mocking purposes.
}

// New is a constructor for Memcached.
func New(exec executor.Executor, config Config) Memcached {
	return Memcached{
		exec:          exec,
		conf:          config,
		isMemcachedUp: netutil.IsListening,
	}

}

func (m Memcached) buildCommand() string {
	return fmt.Sprint(m.conf.PathToBinary,
		" -p ", m.conf.Port,
		" -u ", m.conf.User,
		" -t ", m.conf.NumThreads,
		" -m ", m.conf.MaxMemoryMB,
		" -c ", m.conf.NumConnections)
}

// Launch starts the workload (process or group of processes). It returns a workload
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (m Memcached) Launch() (executor.TaskHandle, error) {
	task, err := m.exec.Execute(m.buildCommand())
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", m.conf.IP, m.conf.Port)
	if !m.isMemcachedUp(address, 5*time.Second) {
		if err := task.Stop(); err != nil {
			log.Errorf("failed to stop memcached instance. Error: %q", err.Error())
		}

		if err := task.Clean(); err != nil {
			log.Errorf("failed to cleanup memcached task. Error: %q", err.Error())
		}

		if err := task.EraseOutput(); err != nil {
			log.Errorf("failed to erase output of memcached task. Error: %q", err.Error())
		}
		return nil, errors.Errorf("failed to connect to memcached instance. Timeout on connection to %q",
			address)
	}
	return task, nil
}

// Name returns human readable name for job.
func (m Memcached) Name() string {
	return name
}
