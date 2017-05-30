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
	"github.com/intelsdi-x/swan/pkg/snap/sessions/caffe"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stressng"
	"github.com/pkg/errors"
)

const (

	// NoneAggressorID is constant to represent "pseudo" aggressor for baselining experiment (running HP workload without aggressor at all).
	NoneAggressorID = "None"

	// CaffeAggressorWithIsolation to run caffe with them same isolation as LLC intesive workloads.
	CaffeAggressorWithIsolation = "caffe-isolated"

	l1dDefaultProcessNumber   = 1
	l1iDefaultProcessNumber   = 1
	l3DefaultProcessNumber    = 1
	membwDefaultProcessNumber = 1
)

var (
	// AggressorsFlag is a comma separated list of aggressors to be run during the experiment.
	AggressorsFlag = conf.NewStringSliceFlag(
		"experiment_be_workloads", "Best Effort workloads that will be run sequentially in colocation with High Priority workload. \n"+
			"# When experiment is run on machine with HyperThreads, user can also add 'stress-ng-cache-l1' to this list. \n"+
			"# When iBench and Stream is available, user can also add 'l1d,l1i,l3,stream' to this list.",
		[]string{NoneAggressorID, stressng.IDCacheL3, stressng.IDMemCpy, stressng.IDStream, caffe.ID},
	)

	theatAggressorsAsService = conf.NewBoolFlag(
		"debug_treat_be_as_service", "Debug only: Best Effort workloads are wrapped in Service flags so that the experiment can track their lifectcle. Default `true` should not be changed without explicit reason.",
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

// AggressorFactory is factory for creating aggressor launchers with local executor
// and isolation.
// Isolation depends on type of aggressor that is given to factory.
type AggressorFactory struct {
	l1AggressorIsolation    isolation.Decorator
	otherAggressorIsolation isolation.Decorator
}

// NewSingleIsolationAggressorFactory returns aggressor launchers factory with local executor.
// Aggressors will have the same isolation.
func NewSingleIsolationAggressorFactory(isolation isolation.Decorator) AggressorFactory {
	return AggressorFactory{
		l1AggressorIsolation:    isolation,
		otherAggressorIsolation: isolation,
	}
}

// NewMultiIsolationAggressorFactory returns factory for aggressor launchers with local executor
// and prepared isolations.
// L1-Data and L1-Instruction cache aggressor will receive l1AggressorIsolation
// Other aggressors will receive otherAggressorIsolation
func NewMultiIsolationAggressorFactory(
	l1AggressorIsolation isolation.Decorator,
	otherAggressorIsolation isolation.Decorator) AggressorFactory {
	return AggressorFactory{
		l1AggressorIsolation:    l1AggressorIsolation,
		otherAggressorIsolation: otherAggressorIsolation,
	}
}

// ExecutorFactoryFunc is a type of function that every call returns new instance of executor with decorators.
type ExecutorFactoryFunc func(isolation.Decorators) (executor.Executor, error)

// Create returns aggressor of chosen type using provided executorFactor to create executor.
func (f AggressorFactory) Create(name string, executorFactory ExecutorFactoryFunc) (executor.Launcher, error) {

	// Zero-value sensitivity.LauncherSessionPair represents baselining.
	if name == NoneAggressorID {
		return nil, nil
	}

	var aggressor executor.Launcher

	decorators := f.getDecorators(name)

	exec, err := executorFactory(decorators)
	errutil.Check(err)

	switch name {
	case l1data.ID:
		aggressor = l1data.New(exec, l1data.DefaultL1dConfig())
	case l1instruction.ID:
		aggressor = l1instruction.New(exec, l1instruction.DefaultL1iConfig())
	case memoryBandwidth.ID:
		aggressor = memoryBandwidth.New(exec, memoryBandwidth.DefaultMemBwConfig())
	case caffe.ID:
		aggressor = caffe.New(exec, caffe.DefaultConfig())
	case CaffeAggressorWithIsolation:
		config := caffe.DefaultConfig()
		config.Name = "Caffe isolated"
		aggressor = caffe.New(exec, config)
	case l3.ID:
		aggressor = l3.New(exec, l3.DefaultL3Config())
	case stream.ID:
		aggressor = stream.New(exec, stream.DefaultConfig())
	case stressng.IDStream:
		aggressor = stressng.NewStream(exec)
	case stressng.IDCacheL1:
		aggressor = stressng.NewCacheL1(exec)
	case stressng.IDCacheL3:
		aggressor = stressng.NewCacheL3(exec)
	case stressng.IDMemCpy:
		aggressor = stressng.NewMemCpy(exec)
	default:
		return nil, errors.Errorf("aggressor %q not found", name)
	}

	if theatAggressorsAsService.Value() {
		aggressor = executor.ServiceLauncher{Launcher: aggressor}
	}

	return aggressor, nil
}

// getDecorators returns decorators that should be applied on executor.
// L1-Data and L1-Instruction cache aggressor receives l1AggressorIsolation
// Other aggressors receive otherAggressorIsolation
func (f AggressorFactory) getDecorators(name string) isolation.Decorators {

	switch name {
	case l1data.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1dProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1dProcessNumber.Value()))
		}
		return decorators
	case l1instruction.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1iProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1iProcessNumber.Value()))
		}
		return decorators
	case l3.ID:
		decorators := isolation.Decorators{f.otherAggressorIsolation}
		if L3ProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L3ProcessNumber.Value()))
		}
		return decorators
	case memoryBandwidth.ID:
		decorators := isolation.Decorators{f.otherAggressorIsolation}
		if MembwProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(MembwProcessNumber.Value()))
		}
		return decorators
	case stressng.IDCacheL1:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		return decorators
	case caffe.ID:
		return isolation.Decorators{}
	case CaffeAggressorWithIsolation:
		return isolation.Decorators{f.otherAggressorIsolation}
	default:
		return isolation.Decorators{f.otherAggressorIsolation}
	}
}

// PrepareAggressors prepare aggressors launchers.
// wrapped by session-less pair using given isolation and executor factory for aggressor workloads.
func PrepareAggressors(l1Isolation, llcIsolation isolation.Decorator, beExecutorFactory ExecutorFactoryFunc) (aggressorPairs []LauncherSessionPair, err error) {
	// Initialize aggressors with BE isolation wrapped as Snap session pairs.
	aggressorFactory := NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)

	for _, aggressorName := range AggressorsFlag.Value() {
		aggressorPair, err := aggressorFactory.Create(aggressorName, beExecutorFactory)
		if err != nil {
			return nil, err
		}

		var launcherSessionPair LauncherSessionPair

		switch aggressorName {

		case NoneAggressorID:
			launcherSessionPair = LauncherSessionPair{}

		case caffe.ID, CaffeAggressorWithIsolation:
			caffeSession, err := caffeinferencesession.NewSessionLauncher(caffeinferencesession.DefaultConfig())
			if err != nil {
				return nil, err
			}
			launcherSessionPair = NewMonitoredLauncher(aggressorPair, caffeSession)
		case l1data.ID, l1instruction.ID, memoryBandwidth.ID, l3.ID, stream.ID, stressng.IDStream, stressng.IDCacheL1, stressng.IDCacheL3, stressng.IDMemCpy:
			launcherSessionPair = NewLauncherWithoutSession(aggressorPair)
		default:
			return nil, errors.Errorf("aggressor %q not found", aggressorName)
		}

		aggressorPairs = append(aggressorPairs, launcherSessionPair)
	}
	return
}
