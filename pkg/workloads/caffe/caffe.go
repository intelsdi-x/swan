package caffe

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/osutil"
	"github.com/intelsdi-x/swan/pkg/swan"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"path"
)

const (
	binaryEnvKey = "SWAN_CAFFE_PATH"
	solverEnvKey = "SWAN_CAFFE_SOLVER_PATH"

	defaultBinaryRelativePath = "tools/caffe"
	defaultSolverRelativePath = "examples/cifar10/cifar10_quick_solver.prototxt"
)

//#!/usr/bin/env sh
//
//TOOLS=./build/tools
//
//$TOOLS/caffe train \
//--solver=examples/cifar10/cifar10_quick_solver.prototxt
//
//# reduce learning rate by factor of 10 after 8 epochs
//$TOOLS/caffe train \
//--solver=examples/cifar10/cifar10_quick_solver_lr1.prototxt \
//--snapshot=examples/cifar10/cifar10_quick_iter_4000.solverstate.h5

func getPathFromEnvOrDefault(envkey string, relativePath string) string {
	return osutil.GetEnvOrDefault(
		envkey, path.Join(swan.GetSwanWorkloadsPath(), relativePath))
}

// Config is a config for the Caffe
type Config struct {
	PathToBinary string
	PathToSolver string
}

// DefaultCaffeConfig is a constructor for caffe.Config with default parameters.
func DefaultConfig() Config {
	return Config{
		PathToBinary: getPathFromEnvOrDefault(binaryEnvKey, defaultBinaryRelativePath),
		PathToSolver: getPathFromEnvOrDefault(solverEnvKey, defaultSolverRelativePath),
	}
}

// Caffe ...
type Caffe struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for Memcached.
func New(exec executor.Executor, config Config) workloads.Launcher {
	return Caffe{
		exec: exec,
		conf: config,
	}

}

func (c Caffe) buildCommand() string {
	return fmt.Sprintf("%s train --solver=%s",
		c.conf.PathToBinary,
		c.conf.PathToSolver)
}

// Launch starts the workload (process or group of processes). It returns a workload
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (c Caffe) Launch() (task executor.TaskHandle, err error) {
	task, err = c.exec.Execute(c.buildCommand())
	if err != nil {
		return nil, err
	}

	return task, err
}

// Name returns human readable name for job.
func (m Caffe) Name() string {
	return "Caffe"
}
