package redis

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"syscall"
	"time"
)

const (
	name                 = "Redis"
	defaultPort          = 6379
	defaultPathToBinary  = "redis-server"
	defaultListenIP      = "0.0.0.0"
	defaultMaxMemory     = "512mb"
	defaultClusterMode   = false
	defaultProtectedMode = false
	defaultTimeout       = 5
	defaultDaemonize     = false
	defaultSudo          = false
	defaultIsolate       = false
)

var (
	// PortFlag return port which will be specified for workload services as endpoints.
	PortFlag          = conf.NewIntFlag("redis_port", "Port of Redis to listen on. (--port)", defaultPort)
	// PathFlag return path to Redis file.
	PathFlag          = conf.NewStringFlag("redis_path", "Path to Redis binary file.", defaultPathToBinary)
	IPFlag            = conf.NewStringFlag("redis_listening_address", "Ip address of interface that Redis will be listening on. It must be actual device address, not '0.0.0.0'.", defaultListenIP)
	MaxMemoryFlag     = conf.NewStringFlag("redis_max_memory", "Maximum memory in Bytes to use for items in bytes. (--maxmemory)", defaultMaxMemory)
	ClusterFlag       = conf.NewBoolFlag("redis_cluster_mode", "Cluster mode parameter.", defaultClusterMode)
	ProtectedModeFlag = conf.NewBoolFlag("redis_protected_mode", "Prodected mode parameter.", defaultProtectedMode)
	TimeoutFlag       = conf.NewIntFlag("redis_timeout", "Maximum wait time for start Redis in seconds.", defaultTimeout)
	DaemonizeFlag     = conf.NewBoolFlag("redis_daemonize", "Daemonize Redis server", defaultDaemonize)
	SudoFlag          = conf.NewBoolFlag("redis_sudo", "Run Redis server in sudo", defaultSudo)
	IsolateFlag       = conf.NewBoolFlag("redis_isolate", "Run Redis server in new namespace pid", defaultIsolate)
)

// Config is a config for the Redis data caching application.
type Config struct {
	PathToBinary  string
	Port          int
	IP            string
	MaxMemory     string
	ClusterMode   bool
	ProtectedMode bool
	Timeout       int
	Daemonize     bool
	Sudo          bool
	Isolate       bool
}

// DefaultConfig is a contructor for Config with default parameters.
func DefaultConfig() Config {
	return Config{
		PathToBinary:  PathFlag.Value(),
		Port:          PortFlag.Value(),
		IP:            IPFlag.Value(),
		MaxMemory:     MaxMemoryFlag.Value(),
		ClusterMode:   ClusterFlag.Value(),
		ProtectedMode: ProtectedModeFlag.Value(),
		Timeout:       TimeoutFlag.Value(),
		Daemonize:     DaemonizeFlag.Value(),
		Sudo:          SudoFlag.Value(),
		Isolate:       IsolateFlag.Value(),
	}
}

// Redis is a launcher for the redis data caching application.
type Redis struct {
	exec      executor.Executor
	conf      Config
	isRedisUp netutil.IsListeningFunction
}

// New is a contructor for Redis.
func New(exec executor.Executor, config Config) Redis {

	return Redis{
		exec:      exec,
		conf:      config,
		isRedisUp: netutil.IsListening,
	}
}

// Launch launches Redis server.
func (r Redis) Launch() (executor.TaskHandle, error) {

	command := r.buildCommand()

	task, err := r.exec.Execute(command)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", task.Address(), r.conf.Port)
	if !r.isRedisUp(address, time.Second*time.Duration(r.conf.Timeout)) {

		if err := task.Stop(); err != nil {
			log.Errorf("failed to stop redis instance. Error: %q", err.Error())
		}

		return nil, errors.Errorf("Failed to connect to redis instance. Timeout on connection to %q !", address)
	}

	return task, nil
}

// Return name of workload.
func (r Redis) String() string {
	return name
}

func (r Redis) buildCommand() string {

	cmd := fmt.Sprint(r.conf.PathToBinary,
		" --port ", r.conf.Port,
		" --bind ", r.conf.IP,
		" --maxmemory ", r.conf.MaxMemory)

	// By default Redis protected mode is enabled.
	if !r.conf.ProtectedMode {
		cmd += " --protected-mode no"
	}

	if r.conf.Daemonize {
		cmd += " --daemonize yes"
	}

	if r.conf.Isolate {

		decorator, err := isolation.NewNamespace(syscall.CLONE_NEWPID)

		if err != nil {
			log.Errorf("Imposible to create namespace decorator: %q", err)
			return ""
		}

		cmd = decorator.Decorate(cmd)
	}

	if r.conf.Sudo {
		cmd = "sudo " + cmd
	}

	return cmd
}
