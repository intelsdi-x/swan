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

package caffe

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/caffe"
	"github.com/pkg/errors"
)

const (
	defaultName         = "Caffe"
	defaultCaffeWrapper = "caffe.sh"
	defaultModel        = "examples/cifar10/cifar10_quick_train_test.prototxt" // relative to caffe binary
	defaultWeights      = "examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5"
	defaultIterations   = 1000000000
	defaultSigintEffect = "stop"
)

var (
	caffeModel   = conf.NewStringFlag("caffe_model", "Path to trained model", defaultModel)
	caffeWeights = conf.NewStringFlag("caffe_weights", "Path to trained weights", defaultWeights)
)

// Config is a config for the Caffe.
type Config struct {
	Name             string
	BinaryPath       string
	LibPath          string
	ModelPath        string
	WeightsPath      string
	IterationsNumber int
	SigintEffect     string

	// Snap APM Collection.
	CollectAPM bool
	SnapConfig caffeinferencesession.Config
	SnapTags   map[string]interface{}
}

// DefaultConfig is a constructor for caffe.Config with default parameters.
func DefaultConfig() Config {
	return Config{
		Name:             defaultName,
		BinaryPath:       defaultCaffeWrapper,
		ModelPath:        caffeModel.Value(),
		WeightsPath:      caffeWeights.Value(),
		IterationsNumber: defaultIterations,
		SigintEffect:     defaultSigintEffect,
		CollectAPM:       snap.UseSnapSessionForWorkloads.Value(),
		SnapConfig:       caffeinferencesession.DefaultConfig(),
		SnapTags:         make(map[string]interface{}),
	}
}

// Caffe is a deep learning framework.
// Implements workload.Launcher.
type Caffe struct {
	exec executor.Executor
	conf Config

	// sessionConstructor is function pointer for UT purposes.
	sessionConstructor func(caffeinferencesession.Config) (snap.SessionLauncher, error)
}

// New is a constructor for Caffe.
func New(exec executor.Executor, config Config) executor.Launcher {
	return Caffe{
		exec:               exec,
		conf:               config,
		sessionConstructor: caffeinferencesession.NewSessionLauncher,
	}

}

func (c Caffe) buildCommand() string {
	return fmt.Sprintf("%s test -model %s -weights %s -iterations %d -sigint_effect %s",
		c.conf.BinaryPath,
		c.conf.ModelPath,
		c.conf.WeightsPath,
		c.conf.IterationsNumber,
		c.conf.SigintEffect)
}

// Launch launches Caffe workload. It's implementation of workload.Launcher interface.
func (c Caffe) Launch() (task executor.TaskHandle, err error) {
	command := c.buildCommand()
	task, err = c.exec.Execute(command)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot launch caffe with command %q", command)
	}

	// TODO(skonefal): Actually, Snap Caffe Collection should launch Caffe Launcher.
	if c.conf.CollectAPM {
		snapLauncher, err := c.sessionConstructor(c.conf.SnapConfig)
		if err != nil {
			task.Stop()
			return nil, err
		}

		snapHandle, err := snapLauncher.LaunchSession(task, c.conf.SnapTags)
		if err != nil {
			task.Stop()
			return nil, err
		}

		return executor.NewClusterTaskHandle(task, []executor.TaskHandle{snapHandle}), nil
	}

	return task, nil
}

// String returns human readable name for job.
func (c Caffe) String() string {
	return c.conf.Name
}
