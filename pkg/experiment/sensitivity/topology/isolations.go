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

package topology

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/pkg/errors"
)

var (
	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("experiment_hp_workload_cpu_count", "Number of CPUs assigned to high priority task. CPUs will be assigned automatically to workloads.", 1)
	beCPUCountFlag = conf.NewIntFlag("experiment_be_workload_cpu_count", "Number of CPUs assigned to best effort task. CPUs will be assigned automatically to workloads.", 1)

	// HpRangeFlag allows to set high priority task cores.
	HpRangeFlag = conf.NewIntSetFlag("experiment_hp_workload_cpu_range", "HP cpuset range (e.g: 0-2). All three 'range' flags must be set to use this policy.", "")
	// BeRangeFlag allows to set best effort task cores with default isolation.
	BeRangeFlag = conf.NewIntSetFlag("experiment_be_workload_l3_cpu_range", "BE cpuset range (e.g: 0-2) for workloads that are targeted as LLC-interfering workloads. All three 'range' flags must be set to use this policy. ", "")
	// BeL1RangeFlag allows to set best effort task cores with L1 cache isolation.
	BeL1RangeFlag = conf.NewIntSetFlag("experiment_be_workload_l1_cpu_range", "BE cpuset range (e.g: 0-2) for workloads that are targeted as L1-interfering workloads. All three 'range' flags must be set to use this policy.", "")
)

type defaultTopology struct {
	HpThreadIDs               isolation.IntSet
	SharingLLCButNotL1Threads isolation.IntSet
	SiblingThreadsToHpThreads topo.ThreadSet
}

// NewIsolations returns HP anb factory of aggressors with applied isolation for BE tasks.
// TODO: needs update for different isolation per cpu
func NewIsolations() (hpIsolation, l1Isolation, llcIsolation isolation.Decorator) {
	if isManualPolicy() {
		llcIsolation = isolation.Taskset{CPUList: BeRangeFlag.Value()}
		l1Isolation = isolation.Taskset{CPUList: BeL1RangeFlag.Value()}
		hpIsolation = isolation.Taskset{CPUList: HpRangeFlag.Value()}

		log.Info("Using Manual Core Placement for workload isolation")
		log.Debugf("HP CPU Threads from flag %q: %s", HpRangeFlag.Name, HpRangeFlag.Value().AsRangeString())
		log.Debugf("BE-LLC CPU Threads from flag %q: %s", BeRangeFlag.Name, BeRangeFlag.Value().AsRangeString())
		log.Debugf("BE-L1  CPU Threads from flag %q: %s", BeL1RangeFlag.Name, BeL1RangeFlag.Value().AsRangeString())
	} else {
		defaultTopology, err := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value())
		errutil.Check(err)
		hpThreads := defaultTopology.HpThreadIDs
		bellcThreads := defaultTopology.SharingLLCButNotL1Threads
		bel1Threads := defaultTopology.SiblingThreadsToHpThreads.AvailableThreads()
		if bel1Threads.Empty() {
			log.Warn("Machine does not support HyperThreads. L1-Cache Best Effort workloads will use LLC threads")
			bel1Threads = bellcThreads
		}

		l1Isolation = isolation.Taskset{CPUList: bel1Threads}
		llcIsolation = isolation.Taskset{CPUList: bellcThreads}
		hpIsolation = isolation.Taskset{CPUList: hpThreads}

		log.Info("Using Automatic Core Placement for workload isolation")
		log.Debugf("HP CPU Threads from flag %q: %v", hpCPUCountFlag.Name, hpThreads)
		log.Debugf("BE-LLC CPU Threads from flag %q: %v", beCPUCountFlag.Name, bellcThreads)
		log.Debugf("BE-L1  CPU Threads from flag %q: %v", beCPUCountFlag.Name, bel1Threads)
	}

	return
}

func isManualPolicy() bool {
	return HpRangeFlag.Value().AsRangeString() != "" &&
		BeRangeFlag.Value().AsRangeString() != "" &&
		BeL1RangeFlag.Value().AsRangeString() != ""
}

func newDefaultTopology(hpCPUCount, beCPUCount int) (defaultTopology, error) {
	var topology defaultTopology
	var err error

	threadSet := sharedCacheThreads()
	topology.HpThreadIDs, err = threadSet.AvailableThreads().Take(hpCPUCount)
	if err != nil {
		return topology, errors.Wrapf(err, "there is not enough cpus to run HP task (%d required)", hpCPUCount)
	}

	// Allocate sibling threads of HP workload to create L1 cache contention
	threadSetOfHpThreads, err := topo.NewThreadSetFromIntSet(topology.HpThreadIDs)
	if err != nil {
		return topology, errors.Wrapf(err, "cannot allocate threads for HP task (threads IDs=%v)", topology.HpThreadIDs)
	}
	topology.SiblingThreadsToHpThreads = getSiblingThreadsOfThreadSet(threadSetOfHpThreads)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(topology.HpThreadIDs)
	topology.SharingLLCButNotL1Threads, err = remaining.Take(beCPUCount)
	if err != nil {
		return topology, errors.Wrapf(err, "cannot allocate remaining threads for BE task (%d required, %d left) - minimum 2 CPUs are required to run experiment", beCPUCount, len(remaining.AsSlice()))
	}

	return topology, nil
}
