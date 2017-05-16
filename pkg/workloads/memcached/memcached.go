// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memcached

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
)

const (
	name = "Memcached"
	// DefaultPort represents default memcached port.
	defaultPort            = 11211
	defaultUser            = "root"
	defaultNumThreads      = 4
	defaultMaxMemoryMB     = 4096
	defaultNumConnections  = 2048
	defaultListenIP        = "127.0.0.1"
	defaultThreadsAffinity = false
)

var (
	// PortFlag returns port which will be specified for workload services as endpoints.
	PortFlag = conf.NewIntFlag("memcached_port", "Port for Memcached to listen on. (-p)", defaultPort)
	// IPFlag returns IP which will be specified for workload services as endpoints.
	IPFlag              = conf.NewStringFlag("memcached_listening_address", "IP address of interface that Memcached will be listening on. It must be actual device address, not '0.0.0.0'.", defaultListenIP)
	userFlag            = conf.NewStringFlag("memcached_user", "Username for Memcached process. (-u)", defaultUser)
	numThreadsFlag      = conf.NewIntFlag("memcached_threads", "Number of threads to use. (-t)", defaultNumThreads)
	threadsAffinityFlag = conf.NewBoolFlag("memcached_threads_affinity", "Threads affinity (-T) (requires memcached patch)", defaultThreadsAffinity)
	maxConnectionsFlag  = conf.NewIntFlag("memcached_connections", "Max simultaneous connections. (-c)", defaultNumConnections)
	maxMemoryMBFlag     = conf.NewIntFlag("memcached_max_memory", "Maximum memory in MB to use for items in megabytes. (-m)", defaultMaxMemoryMB)
)

// Config is a config for the memcached data caching application v 1.4.25.
// memcached 1.4.25, supported options:
// -p <num>      TCP port number to listen on (default: 11211)
// -s <file>     UNIX socket path to listen on (disables network support)
// -u <username> assume identity of <username> (only when run as root)
// -m <num>      max memory to use for items in megabytes (default: 64 MB)
// -c <num>      max simultaneous connections (default: 1024)
//-v            verbose (print errors/warnings while in event loop)
//-vv           very verbose (also print client commands/responses)
//-vvv          extremely verbose (also print internal state transitions)
//-t <num>      number of threads to use (default: 4)
type Config struct {
	PathToBinary    string
	Port            int
	User            string
	NumThreads      int
	ThreadsAffinity bool
	MaxMemoryMB     int
	NumConnections  int
	IP              string
}

// DefaultMemcachedConfig is a constructor for MemcachedConfig with default parameters.
func DefaultMemcachedConfig() Config {
	return Config{
		PathToBinary:    "memcached",
		Port:            PortFlag.Value(),
		User:            userFlag.Value(),
		NumThreads:      numThreadsFlag.Value(),
		ThreadsAffinity: threadsAffinityFlag.Value(),
		MaxMemoryMB:     maxMemoryMBFlag.Value(),
		NumConnections:  maxConnectionsFlag.Value(),
		IP:              IPFlag.Value(),
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
	if config.IP == "0.0.0.0" {
		log.Panic("Memcached has to listen on actual device address, not '0.0.0.0'")
	}
	return Memcached{
		exec:          exec,
		conf:          config,
		isMemcachedUp: netutil.IsListening,
	}

}

func (m Memcached) buildCommand() string {
	cmd := fmt.Sprint(m.conf.PathToBinary,
		" -p ", m.conf.Port,
		" -u ", m.conf.User,
		" -t ", m.conf.NumThreads,
		" -m ", m.conf.MaxMemoryMB,
		" -c ", m.conf.NumConnections)
	if m.conf.ThreadsAffinity {
		cmd += " -T"
	}
	return cmd
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

		return nil, errors.Errorf("failed to connect to memcached instance. Timeout on connection to %q",
			address)
	}
	return task, nil
}

// Name returns human readable name for job.
func (m Memcached) Name() string {
	return name
}
