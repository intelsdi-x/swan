package stream

import (
	"fmt"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID   = "stream"
	name = "Stream 100M"
)

// PathFlag represents stream path flag.
// You can override it to point binary of stress with different problem size
// eg. -stream_path=low-level-aggressors/stresm.50M.
var PathFlag = conf.NewStringFlag(
	"stream_path",
	"Path to stream binary",
	"stream.100M",
)

// Config is a struct for stream aggressor configuration.
type Config struct {
	Path string
	// 0 (default) means use all available threads.
	// (https://gcc.gnu.org/onlinedocs/libgomp/OMP_005fNUM_005fTHREADS.html#OMP_005fNUM_005fTHREADS).
	NumThreads uint
}

// DefaultConfig is a constructor for l1d aggressor Config with default parameters.
func DefaultConfig() Config {
	return Config{
		Path: PathFlag.Value(),
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
	return fmt.Sprintf("sh -c 'OMP_NUM_THREADS=%d %s'", l.conf.NumThreads, l.conf.Path)
}

// Launch starts a workload.
func (l stream) Launch() (executor.TaskHandle, error) {
	return l.exec.Execute(l.buildCommand())
}

// Name returns human readable name for job.
func (l stream) Name() string {
	return name
}
