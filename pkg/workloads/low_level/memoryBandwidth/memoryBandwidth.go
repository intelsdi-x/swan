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

package memoryBandwidth

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID              = "membw"
	name            = "memBW"
	defaultDuration = 86400 * time.Second
)

// Config is a struct for MemBw aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultMemBwConfig is a constructor for memBw aggressor Config with default parameters.
func DefaultMemBwConfig() Config {
	return Config{
		Path:     "memBw",
		Duration: defaultDuration,
	}
}

// memBw is a launcher for memBw aggressor.
type memBw struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for memBw aggressor.
func New(exec executor.Executor, config Config) executor.Launcher {
	return memBw{
		exec: exec,
		conf: config,
	}
}

func (m memBw) buildCommand() string {
	return fmt.Sprintf("%s %d", m.conf.Path, int(m.conf.Duration.Seconds()))
}

func (m memBw) verifyConfiguration() error {
	if m.conf.Duration.Seconds() <= 0 {
		return errors.Errorf("launcher configuration is invalid. `duration` value(%d) is lower/equal than/to 0",
			int(m.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (m memBw) Launch() (executor.TaskHandle, error) {
	if err := m.verifyConfiguration(); err != nil {
		return nil, err
	}
	return m.exec.Execute(m.buildCommand())
}

// String returns human readable name for job.
func (m memBw) String() string {
	return name
}
