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

package stressng

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
)

// StressngCustomArguments custom argument to run stress-ng with.
var StressngCustomArguments = conf.NewStringFlag("stressng_custom_arguments", "Custom arguments to stress-ng", "")

// StressngStreamProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngStreamProcessNumber = conf.NewIntFlag("stressng_stream_process_number", "Number of aggressors to be run", 1)

// StressngCacheL1ProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngCacheL1ProcessNumber = conf.NewIntFlag("stressng_cache_l1_process_number", "Number of aggressors to be run", 1)

// StressngCacheL3ProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngCacheL3ProcessNumber = conf.NewIntFlag("stressng_cache_l3_process_number", "Number of aggressors to be run", 1)

// StressngMemCpyProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngMemCpyProcessNumber = conf.NewIntFlag("stressng_memcpy_process_number", "Number of aggressors to be run", 1)

// l1d is a launcher for stress-ng aggressor.
type stressng struct {
	executor  executor.Executor
	arguments string
	name      string
}

// New is a constructor for stress-ng aggressor.
func New(executor executor.Executor, name, arguments string) executor.Launcher {
	return stressng{
		executor:  executor,
		arguments: arguments,
		name:      name,
	}
}

// NewCustom constructor for specifc run of stress-ng.
func NewCustom(executor executor.Executor) executor.Launcher {
	return New(executor, fmt.Sprintf("stress-ng-custom %s", StressngCustomArguments.Value()), StressngCustomArguments.Value())
}

// NewStream constructor for stream based run of stress-ng.
func NewStream(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-stream", fmt.Sprintf("--stream=%d", StressngStreamProcessNumber.Value()))
}

// NewCacheL1 constructor for cache L1 run of stress-ng.
func NewCacheL1(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-cache-l1", fmt.Sprintf("--cache=%d --cache-level=1", StressngCacheL1ProcessNumber.Value()))
}

// NewCacheL3 constructor for cache L3 run of stress-ng.
func NewCacheL3(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-cache-l3", fmt.Sprintf("--cache=%d --cache-level=3", StressngCacheL3ProcessNumber.Value()))
}

// NewMemCpy constructor for memcpy stressor run of stress-ng.
func NewMemCpy(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-memcpy", fmt.Sprintf("--memcpy=%d", StressngMemCpyProcessNumber.Value()))
}

// Launch starts a workload.
func (s stressng) Launch() (executor.TaskHandle, error) {
	return s.executor.Execute(fmt.Sprintf("stress-ng %s", s.arguments))
}

// String returns readable name.
func (s stressng) String() string {
	return s.name
}
