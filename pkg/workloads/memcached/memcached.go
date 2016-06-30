package memcached

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"path"
)

const (
	name = "Memcached"
	// DefaultPort represents default memcached port.
	DefaultPort           = 11211
	defaultUser           = "memcached"
	defaultNumThreads     = 4
	defaultMaxMemoryMB    = 64
	defaultNumConnections = 1024
	defaultListenIP       = "127.0.0.1"
)

// PathFlag represents memcached path flag.
var PathFlag = conf.NewFileFlag(
	"memcached_path",
	"Path to memcached binary",
	path.Join(fs.GetSwanWorkloadsPath(), "data_caching/memcached/memcached-1.4.25/build/memcached"),
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
		PathToBinary:   PathFlag.Value(),
		Port:           DefaultPort,
		User:           defaultUser,
		NumThreads:     defaultNumThreads,
		MaxMemoryMB:    defaultMaxMemoryMB,
		NumConnections: defaultNumConnections,
		IP:             defaultListenIP,
	}
}

type dialTimeoutFunc func(address string, timeout time.Duration) bool

func tryConnect(address string, timeout time.Duration) bool {
	retries := 5
	sleepTime := time.Duration(
		timeout.Nanoseconds() / int64(retries))
	connected := false
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}

// Memcached is a launcher for the memcached data caching application v 1.4.25.
type Memcached struct {
	exec       executor.Executor
	conf       Config
	tryConnect dialTimeoutFunc // For mocking purposes.
}

// New is a constructor for Memcached.
func New(exec executor.Executor, config Config) Memcached {
	return Memcached{
		exec:       exec,
		conf:       config,
		tryConnect: tryConnect,
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
	if !m.tryConnect(address, 5*time.Second) {
		err := fmt.Errorf("Failed to connect to memcached instance. Timeout on connection to %s",
			address)
		log.Error(err)

		err1 := task.Stop()
		if err1 != nil {
			log.Error("Failed to stop memcached instance. Error: " + err1.Error())
		}
		err1 = task.Clean()
		if err1 != nil {
			log.Error("Failed to cleanup memcached task. Error: " + err1.Error())
		}
		return nil, err
	}
	return task, nil
}

// Name returns human readable name for job.
func (m Memcached) Name() string {
	return name
}
