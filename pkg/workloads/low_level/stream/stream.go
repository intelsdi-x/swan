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

package stream

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	name = "Stream 100M"
)

// StreamThreadNumberFlag is a flag that allows to control number of stream aggressor's threads. 0 (default) means use all available threads.
// (https://gcc.gnu.org/onlinedocs/libgomp/OMP_005fNUM_005fTHREADS.html#OMP_005fNUM_005fTHREADS).
var StreamThreadNumberFlag = conf.NewIntFlag("experiment_be_stream_thread_number", "Number of threads that stream aggressor is going to launch. Default value (0) will launch one thread per cpu.", 0)

// Config is a struct for stream aggressor configuration.
type Config struct {
	Path       string
	NumThreads uint
}

// DefaultConfig is a constructor for l1d aggressor Config with default parameters.
func DefaultConfig() Config {
	return Config{
		Path:       "stream.100M",
		NumThreads: uint(StreamThreadNumberFlag.Value()),
	}
}

type stream struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for stream benchmark aggressor.
// STREAM: Sustainable Memory Bandwidth in High Performance Computers
// https://www.cs.virginia.edu/stream/
//
// Stream Benchmark working set is 100 million of elements (type double).
// Working set size should be more than 4x the size of sum of all last-level cache used in the run
// If you need more consider rebuilding stream with STREAM_ARRAY_SIZE adjusted accordingly.
// Check stream.c "Instructions" for more details.
func New(exec executor.Executor, config Config) executor.Launcher {
	return stream{
		exec: exec,
		conf: config,
	}
}

func (l stream) buildCommand() string {
	return fmt.Sprintf("env OMP_NUM_THREADS=%d %s", l.conf.NumThreads, l.conf.Path)
}

// Launch starts a workload.
func (l stream) Launch() (executor.TaskHandle, error) {
	return l.exec.Execute(l.buildCommand())
}

// String returns human readable name for job.
func (l stream) String() string {
	return name
}
