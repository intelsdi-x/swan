package workloads

import "github.com/intelsdi-x/swan/pkg/executor"

const (
	memcachedBinaryName = "memcached"
)

// Memcached is a launcher for memcached data caching application.
type Memcached struct {
	exec         executor.Executor
	pathToBinary string
	user         string
	threadsNum   int
	connNum      int
}

// NewMemcached is a constructor for Memcached.
func NewMemcached(exec executor.Executor, pathToBinary string,
	user string, threadsNum int, connNum int) Memcached {
	return Memcached{
		exec:         exec,
		pathToBinary: pathToBinary,
		user:         user,
		threadsNum:   threadsNum,
		connNum:      connNum,
	}
}

func (m Memcached) buildCommand() string {
	return m.pathToBinary + memcachedBinaryName +
		" -u " + m.user +
		" -t " + string(m.threadsNum) +
		" -c " + string(m.connNum) +
		" -m mem"
}

// Launch starts the workload (process or group of processes). It returns a workload
// represented as a Task instance.
// Error is returned when Launcher is unable to start a job.
func (m Memcached) Launch() (executor.Task, error) {
	return m.exec.Execute(m.buildCommand())
}
