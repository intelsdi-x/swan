package workloads

import "github.com/intelsdi-x/swan/pkg/executor"

const (
	memcachedBinaryName = "memcached"
)

type Memcached struct {
	exec         executor.Executor
	pathToBinary string
	user         string
	threadsNum   int
	connNum      int
}

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
		" -t " + m.threadsNum +
		" -c " + m.connNum +
		" -m mem"
}

func (m Memcached) Launch() (executor.Task, error) {
	return m.exec.Execute(m.buildCommand())
}
