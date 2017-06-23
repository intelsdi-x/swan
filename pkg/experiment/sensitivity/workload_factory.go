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

package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stressng"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	"github.com/pkg/errors"
)

const (
	// NoneAggressorID is constant to represent "pseudo" aggressor for baselining experiment (running HP workload without aggressor at all).
	NoneAggressorID = "None"

	// High Priority workloads.

	// Memcached workload.
	Memcached = "memcached"
	// Specjbb workload.
	Specjbb = "specjbb"

	// Best Effort workloads.
	caffeWorkload              = "caffe"
	caffeWorkloadWithIsolation = "caffe-isolated"

	l1d      = "l1d"
	l1i      = "l1i"
	llc      = "l3"
	membw    = "membw"
	streambw = "stream"

	stressngL1     = "stress-ng-cache-l1"
	strssngL3      = "stress-ng-cache-l3"
	stressngMemcpy = "stress-ng-memcpy"
	stressngStream = "stress-ng-stream"

	l1dDefaultProcessNumber   = 1
	l1iDefaultProcessNumber   = 1
	l3DefaultProcessNumber    = 1
	membwDefaultProcessNumber = 1
)

var (
	// AggressorsFlag is a comma separated list of aggressors to be run during the experiment.
	AggressorsFlag = conf.NewStringSliceFlag(
		"experiment_be_workloads", "Best Effort workloads that will be run sequentially in colocation with High Priority workload.\n"+
			"When experiment is run on machine with HyperThreads, user can also add 'stress-ng-cache-l1' to this list.\n"+
			"When iBench and Stream is available, user can also add 'l1d,l1i,l3,stream' to this list.",
		[]string{NoneAggressorID, strssngL3, stressngMemcpy, stressngStream, caffeWorkload},
	)

	treatAggressorsAsService = conf.NewBoolFlag(
		"debug_treat_be_as_service", "Debug only: Best Effort workloads are wrapped in Service flags so that the experiment can track their lifecycle. Default `true` should not be changed without explicit reason.",
		true)

	// L1dProcessNumber represents number of L1 data cache aggressor processes to be run
	L1dProcessNumber = conf.NewIntFlag(
		"experiment_be_l1d_processes_number",
		"Number of L1 data cache best effort processes to be run",
		l1dDefaultProcessNumber,
	)

	// L1iProcessNumber represents number of L1 instruction cache aggressor processes to be run
	L1iProcessNumber = conf.NewIntFlag(
		"experiment_be_l1i_processes_number",
		"Number of L1 instruction cache best effort processes to be run",
		l1iDefaultProcessNumber,
	)

	// L3ProcessNumber represents number of L3 data cache aggressor processes to be run
	L3ProcessNumber = conf.NewIntFlag(
		"experiment_be_l3_processes_number",
		"Number of L3 data cache best effort processes to be run",
		l3DefaultProcessNumber,
	)

	// MembwProcessNumber represents number of membw aggressor processes to be run
	MembwProcessNumber = conf.NewIntFlag(
		"experiment_be_membw_processes_number",
		"Number of membw best effort processes to be run",
		membwDefaultProcessNumber,
	)
)

// WorkloadFactory is creator for High Priority and Best Effort workloads with
// default or custom isolation.
type WorkloadFactory struct {
	executorFactory ExecutorFactory

	hpIsolation isolation.Decorator
	l1Isolation isolation.Decorator
	l3Isolation isolation.Decorator
}

// NewDefaultWorkloadFactory returns factory that would create Workloads on Kubernetes executor
// or Local executor, depending on flags.
func NewDefaultWorkloadFactory() WorkloadFactory {
	return NewWorkloadFactory(NewExecutorFactory())
}

// NewWorkloadFactory creates new instance of WorkloadFactory.
func NewWorkloadFactory(executorFactory ExecutorFactory) WorkloadFactory {
	hpIsolation, l1Isolation, l3Isolation := GetWorkloadsIsolations()
	return NewWorkloadFactoryWithIsolation(executorFactory, hpIsolation, l1Isolation, l3Isolation)
}

// NewWorkloadFactoryWithIsolation creates new workload factory with custom isolation.
func NewWorkloadFactoryWithIsolation(
	factory ExecutorFactory,
	hpIsolation isolation.Decorator,
	l1Isolation isolation.Decorator,
	l3Isolation isolation.Decorator) WorkloadFactory {
	return WorkloadFactory{
		executorFactory: factory,
		hpIsolation:     hpIsolation,
		l1Isolation:     l1Isolation,
		l3Isolation:     l3Isolation,
	}
}

// BuildDefaultHighPriorityLauncher builds High Priority workload launcher with predefined isolation.
func (factory *WorkloadFactory) BuildDefaultHighPriorityLauncher(
	workloadName string, tags snap.Tags) (launcher executor.Launcher, err error) {
	return factory.BuildHighPriorityLauncherWithIsolation(workloadName, factory.hpIsolation, tags)
}

// BuildHighPriorityLauncherWithIsolation builds High Priority launcher with provided isolation.
func (factory *WorkloadFactory) BuildHighPriorityLauncherWithIsolation(
	workloadName string,
	decorators isolation.Decorator,
	tags snap.Tags) (launcher executor.Launcher, err error) {
	return factory.createHighPriorityWorkload(workloadName, decorators, tags)
}

// BuildDefaultBestEffortLauncher builds Best Effort workload launcher with predefined isolation.
func (factory *WorkloadFactory) BuildDefaultBestEffortLauncher(
	workloadName string,
	tags snap.Tags) (launcher executor.Launcher, err error) {
	return factory.BuildBestEffortLauncherWithIsolation(workloadName, factory.getDefaultBestEffortIsolation(workloadName), tags)
}

// BuildBestEffortLauncherWithIsolation builds Best Effort launcher with provided isolation.
func (factory *WorkloadFactory) BuildBestEffortLauncherWithIsolation(
	workloadName string,
	isolation isolation.Decorator,
	tags snap.Tags) (launcher executor.Launcher, err error) {
	return factory.createBestEffortWorkload(workloadName, isolation, tags)
}

func (factory *WorkloadFactory) createHighPriorityWorkload(
	name string,
	isolation isolation.Decorator,
	tags snap.Tags) (executor.Launcher, error) {

	exec, err := factory.executorFactory.BuildHighPriorityExecutor(isolation)
	if err != nil {
		return nil, err
	}

	switch name {
	case Memcached:
		return executor.NewServiceLauncher(memcached.New(exec, memcached.DefaultMemcachedConfig())), nil
	case Specjbb:
		return executor.NewServiceLauncher(specjbb.NewBackend(exec, specjbb.DefaultSPECjbbBackendConfig())), nil
	default:
		return nil, errors.Errorf("unknown high priority task %q", name)
	}
}

func (factory *WorkloadFactory) createBestEffortWorkload(
	name string,
	isolation isolation.Decorator,
	tags snap.Tags) (executor.Launcher, error) {

	if name == NoneAggressorID {
		return nil, nil
	}

	var workload executor.Launcher
	additionalDecorators := factory.getBestEffortAdditionalDecorators(name)
	exec, err := factory.executorFactory.BuildBestEffortExecutor(isolation, additionalDecorators)
	if err != nil {
		return nil, err
	}

	// Best Effort workloads.
	switch name {
	case l1d:
		workload = l1data.New(exec, l1data.DefaultL1dConfig())
	case l1i:
		workload = l1instruction.New(exec, l1instruction.DefaultL1iConfig())
	case membw:
		workload = memoryBandwidth.New(exec, memoryBandwidth.DefaultMemBwConfig())
	case caffeWorkload:
		config := caffe.DefaultConfig()
		config.SnapTags = tags
		workload = caffe.New(exec, config)
	case caffeWorkloadWithIsolation:
		config := caffe.DefaultConfig()
		config.Name = "Caffe isolated"
		config.SnapTags = tags
		workload = caffe.New(exec, config)
	case llc:
		workload = l3.New(exec, l3.DefaultL3Config())
	case streambw:
		workload = stream.New(exec, stream.DefaultConfig())
	case stressngL1:
		workload = stressng.NewCacheL1(exec)
	case strssngL3:
		workload = stressng.NewCacheL3(exec)
	case stressngMemcpy:
		workload = stressng.NewMemCpy(exec)
	case stressngStream:
		workload = stressng.NewStream(exec)
	default:
		return nil, errors.Errorf("unknown best effort task %q", name)
	}

	if treatAggressorsAsService.Value() {
		workload = executor.ServiceLauncher{Launcher: workload}
	}

	return workload, nil
}

func (factory *WorkloadFactory) getDefaultBestEffortIsolation(workloadName string) isolation.Decorator {
	switch workloadName {
	case l1d:
		return isolation.Decorators{factory.l1Isolation}
	case l1i:
		return isolation.Decorators{factory.l1Isolation}
	case stressngL1:
		return isolation.Decorators{factory.l1Isolation}
	case caffeWorkload:
		return isolation.Decorators{}
	case caffeWorkloadWithIsolation:
		return isolation.Decorators{factory.l3Isolation}
	default:
		return isolation.Decorators{factory.l3Isolation}
	}
}

func (factory *WorkloadFactory) getBestEffortAdditionalDecorators(workloadName string) isolation.Decorator {
	switch workloadName {
	case l1d:
		if L1dProcessNumber.Value() != 1 {
			return executor.NewParallel(L1dProcessNumber.Value())
		}
	case l1i:
		if L1iProcessNumber.Value() != 1 {
			return executor.NewParallel(L1iProcessNumber.Value())
		}
	case llc:
		if L3ProcessNumber.Value() != 1 {
			return executor.NewParallel(L3ProcessNumber.Value())
		}
	case membw:
		if MembwProcessNumber.Value() != 1 {
			return executor.NewParallel(MembwProcessNumber.Value())
		}
	}

	return isolation.Decorators{}
}
