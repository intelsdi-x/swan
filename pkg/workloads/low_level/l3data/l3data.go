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

package l3data

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID              = "l3d"
	name            = "L3 Data"
	defaultDuration = 86400 * time.Second
)

// Config is a struct for l3 aggressor configuration.
type Config struct {
	Path     string
	Duration time.Duration
}

// DefaultL3Config is a constructor for l3 aggressor Config with default parameters.
func DefaultL3Config() Config {
	return Config{
		Path:     "l3",
		Duration: defaultDuration,
	}
}

// l3 is a launcher for l3 aggressor.
type l3 struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for l3 aggressor.
func New(exec executor.Executor, config Config) executor.Launcher {
	return l3{
		exec: exec,
		conf: config,
	}
}

func (l l3) buildCommand() string {
	return fmt.Sprintf("%s %d", l.conf.Path, int(l.conf.Duration.Seconds()))
}

func (l l3) verifyConfiguration() error {
	if l.conf.Duration.Seconds() <= 0 {
		return errors.Errorf("launcher configuration is invalid. `duration` value(%v) is lower/equal than/to 0",
			int(l.conf.Duration.Seconds()))
	}
	return nil
}

// Launch starts a workload.
// It returns a workload represented as a Task instance.
// Error is returned when Launcher is unable to start a job or when configuration is invalid.
func (l l3) Launch() (executor.TaskHandle, error) {
	if err := l.verifyConfiguration(); err != nil {
		return nil, err
	}
	return l.exec.Execute(l.buildCommand())
}

// Name returns human readable name for job.
func (l l3) Name() string {
	return name
}
