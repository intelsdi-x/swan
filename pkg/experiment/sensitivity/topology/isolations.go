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
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)

	// For manually provided isolation policy.
	hpRangeFlag   = conf.NewIntSetFlag("hp_range", "HP cpuset range", "")
	beRangeFlag   = conf.NewIntSetFlag("be_range", "BE cpuset range", "")
	beL1RangeFlag = conf.NewIntSetFlag("be_l1_range", "BE for l1 aggressors cpuset range", "")
)

type defaultTopology struct {
	hpThreadIDs               isolation.IntSet
	sharingLLCButNotL1Threads isolation.IntSet
	siblingThreadsToHpThreads topo.ThreadSet
}

// NewIsolations returns HP anb factory of aggressors with applied isolation for BE tasks.
// TODO: needs update for different isolation per cpu
func NewIsolations() (hpIsolation, l1Isolation, llcIsolation isolation.Decorator) {
	if isManualPolicy() {
		llcIsolation = isolation.Taskset{CPUList: beRangeFlag.Value()}
		l1Isolation = isolation.Taskset{CPUList: beL1RangeFlag.Value()}
		hpIsolation = isolation.Taskset{CPUList: hpRangeFlag.Value()}
	} else {
		defaultTopology, err := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value())
		errutil.Check(err)
		l1Isolation = isolation.Taskset{CPUList: defaultTopology.siblingThreadsToHpThreads.AvailableThreads()}
		llcIsolation = isolation.Taskset{CPUList: defaultTopology.sharingLLCButNotL1Threads}
		hpIsolation = isolation.Taskset{CPUList: defaultTopology.hpThreadIDs}
	}
	return
}

func isManualPolicy() bool {
	return hpRangeFlag.Value().AsRangeString() != "" && beRangeFlag.Value().AsRangeString() != ""
}

func newDefaultTopology(hpCPUCount, beCPUCount int) (defaultTopology, error) {
	var topology defaultTopology
	var err error

	threadSet := sharedCacheThreads()
	topology.hpThreadIDs, err = threadSet.AvailableThreads().Take(hpCPUCount)
	if err != nil {
		return topology, errors.Wrapf(err, "there is not enough cpus to run HP task (%d required)", hpCPUCount)
	}

	// Allocate sibling threads of HP workload to create L1 cache contention
	threadSetOfHpThreads, err := topo.NewThreadSetFromIntSet(topology.hpThreadIDs)
	if err != nil {
		return topology, errors.Wrapf(err, "cannot allocate threads for HP task (threads IDs=%v)", topology.hpThreadIDs)
	}
	topology.siblingThreadsToHpThreads = getSiblingThreadsOfThreadSet(threadSetOfHpThreads)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(topology.hpThreadIDs)
	topology.sharingLLCButNotL1Threads, err = remaining.Take(beCPUCount)
	if err != nil {
		return topology, errors.Wrapf(err, "cannot allocate remaining threads for BE task (%d required, %d left) - minimum 2 CPUs are required to run experiment", beCPUCount, len(remaining.AsSlice()))
	}

	return topology, nil
}
