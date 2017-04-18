package topology

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/pkg/errors"
)

var (
	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task. It should not be used together with `Hp/BeRangeFlag`", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task. It should not be used together with `Hp/BeRangeFlag`", 1)

	// HpRangeFlag allows to set high priority task cores.
	HpRangeFlag = conf.NewIntSetFlag("hp_range", "HP cpuset range (e.g: 0-2). It should not be used together with 'hp/beCPUCountFlag'. ", "")
	// BeRangeFlag allows to set best effort task cores with default isolation.
	BeRangeFlag = conf.NewIntSetFlag("be_range", "BE cpuset range (e.g: 0-2). It should not be used together with 'hp/beCPUCountFlag'. ", "")
	// BeL1RangeFlag allows to set best effort task cores with L1 cache isolation.
	BeL1RangeFlag = conf.NewIntSetFlag("be_l1_range", "BE for l1 aggressors cpuset range (e.g: 0-2). It should not be used together with 'hp/beCPUCountFlag'. ", "")
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
	} else {
		defaultTopology, err := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value())
		errutil.Check(err)
		l1Isolation = isolation.Taskset{CPUList: defaultTopology.SiblingThreadsToHpThreads.AvailableThreads()}
		llcIsolation = isolation.Taskset{CPUList: defaultTopology.SharingLLCButNotL1Threads}
		hpIsolation = isolation.Taskset{CPUList: defaultTopology.HpThreadIDs}
	}
	return
}

func isManualPolicy() bool {
	return HpRangeFlag.Value().AsRangeString() != "" && BeRangeFlag.Value().AsRangeString() != ""
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
