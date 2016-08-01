package memcached

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
	"path"
)

const (
	name               = "Memcached"
	defaultMaxMemoryMB = 64
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
	Path           string `help:"Path to memcached binary" type:"file" defaultFromField:"defaultPath"`
	defaultPath    string
	Port           int    `help:"Port for memcached to listen on. (-p)" default:"11211"`
	User           string `help:"Username for memcached process (-u)" default:"memcached"`
	NumThreads     int    `help:"Number of threads for mutilate (-t)" name:"Threads" default:"4"`
	NumConnections int    `help:"Number of maximum connections for mutilate (-c)" name:"Connections" default:"1024"`
	IP             string `help:"IP of interface memcached is listening on." type:"ip" default:"127.0.0.1"`

	MaxMemoryMB int
	flagPrefix  string
}

var defaultConfig = Config{
	MaxMemoryMB: defaultMaxMemoryMB,
	defaultPath: path.Join(fs.GetSwanWorkloadsPath(), "data_caching/memcached/memcached-1.4.25/build/memcached"),
	flagPrefix:  name,
}

func init() {
	conf.Process(&defaultConfig)
}

// DefaultConfig is a constructor for MemcachedConfig with default parameters.
func DefaultConfig() Config {
	conf.Process(&defaultConfig)
	return defaultConfig
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
	return fmt.Sprint(m.conf.Path,
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
		err := errors.Errorf("failed to connect to memcached instance. Timeout on connection to %q",
			address)

		err1 := task.Stop()
		if err1 != nil {
			log.Errorf("failed to stop memcached instance. Error: %q", err1.Error())
		}
		err1 = task.Clean()
		if err1 != nil {
			log.Errorf("failed to cleanup memcached task. Error: %q", err1.Error())
		}
		return nil, err
	}
	return task, nil
}

// Name returns human readable name for job.
func (m Memcached) Name() string {
	return name
}
