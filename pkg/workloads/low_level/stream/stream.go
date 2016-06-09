package stream

import (
	"fmt"

	"path"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
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
	path.Join(fs.GetSwanWorkloadsPath(), "low-level-aggressors/stream.100M"),
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

// stream is a launcher for stream aggressor.
// Note: By default is uses 100M array size (problem size). It is enough to saturate about 200MB L3
// cache over whole platform.  If you need more consider rebuilding stream with STREAM_ARRAY_SIZE
// adjusted accordingly. Check stream.c "Instructions" for more details.
type stream struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for stream aggressor.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return stream{
		exec: exec,
		conf: config,
	}
}

func (l stream) buildCommand() string {
	return fmt.Sprintf("OMP_NUM_THREADS=%d %s", l.conf.NumThreads, l.conf.Path)
}

// Launch starts a workload.
func (l stream) Launch() (executor.TaskHandle, error) {
	return l.exec.Execute(l.buildCommand())
}

// Name returns human readable name for job.
func (l stream) Name() string {
	return name
}
