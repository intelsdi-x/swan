package caffe

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/osutil"
	"github.com/intelsdi-x/swan/pkg/swan"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"os"
	"path"
)

const (
	binaryEnvKey  = "SWAN_CAFFE_BINARY_PATH"
	solverEnvKey  = "SWAN_CAFFE_SOLVER_PATH"
	workdirEnvKey = "SWAN_CAFFE_WORKING_DIR_PATH"

	defaultBinaryRelativePath  = "deep_learning/caffe/caffe_src/build/tools/caffe"
	defaultSolverRelativePath  = "deep_learning/caffe/caffe_src/examples/cifar10/cifar10_quick_solver.prototxt"
	defaultWorkdirRelativePath = "deep_learning/caffe/caffe_src/"
)

func getPathFromEnvOrDefault(envkey string, relativePath string) string {
	return osutil.GetEnvOrDefault(
		envkey, path.Join(swan.GetSwanWorkloadsPath(), relativePath))
}

// Config is a config for the Caffe
type Config struct {
	BinaryPath  string
	SolverPath  string
	WorkdirPath string
}

// DefaultConfig is a constructor for caffe.Config with default parameters.
func DefaultConfig() Config {
	return Config{
		BinaryPath:  getPathFromEnvOrDefault(binaryEnvKey, defaultBinaryRelativePath),
		SolverPath:  getPathFromEnvOrDefault(solverEnvKey, defaultSolverRelativePath),
		WorkdirPath: getPathFromEnvOrDefault(workdirEnvKey, defaultWorkdirRelativePath),
	}
}

// Caffe is a deep learning framework
// Implements workload.Launcher
type Caffe struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for Caffe.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return Caffe{
		exec: exec,
		conf: config,
	}

}

func (c Caffe) buildCommand() string {
	return fmt.Sprintf("%s train --solver=%s",
		c.conf.BinaryPath,
		c.conf.SolverPath)
}

// Launch launches Caffe workload. It's implementation of workload.Launcher interface
// Caffe needs to run from it's own working directory, because
// solver look for relative paths when dealing with training and testing
// sets.
func (c Caffe) Launch() (task executor.TaskHandle, err error) {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer popWorkingDir(currentWorkingDir)

	err = os.Chdir(c.conf.WorkdirPath)
	if err != nil {
		return nil, err
	}

	task, err = c.exec.Execute(c.buildCommand())
	if err != nil {
		return nil, err
	}

	return task, err
}

func popWorkingDir(workdir string) {
	os.Chdir(workdir)
}

// Name returns human readable name for job.
func (c Caffe) Name() string {
	return "Caffe"
}
