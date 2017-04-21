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

package l1data

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID              = "l1d"
	name            = "L1 Data"
	defaultDuration = 86400 * time.Second
)

// Config is a struct for l1d aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultL1dConfig is a constructor for l1d aggressor Config with default parameters.
func DefaultL1dConfig() Config {
	return Config{
		Path:     "l1d",
		Duration: defaultDuration,
	}
}

// l1d is a launcher for l1d aggressor.
type l1d struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for l1d aggressor.
func New(exec executor.Executor, config Config) executor.Launcher {
	return l1d{
		exec: exec,
		conf: config,
	}
}

func (l l1d) buildCommand() string {
	return fmt.Sprintf("%s %d", l.conf.Path, int(l.conf.Duration.Seconds()))
}

func (l l1d) verifyConfiguration() error {
	if l.conf.Duration.Seconds() <= 0 {
		return fmt.Errorf("launcher configuration is invalid. `duration` value(%d) is lower/equal than/to 0",
			int(l.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (l l1d) Launch() (executor.TaskHandle, error) {
	if err := l.verifyConfiguration(); err != nil {
		return nil, err
	}
	return l.exec.Execute(l.buildCommand())
}

// Name returns human readable name for job.
func (l l1d) Name() string {
	return name
}
