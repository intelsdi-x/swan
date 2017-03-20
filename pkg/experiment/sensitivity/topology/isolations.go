package topology

import (
	"strconv"
	"strings"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/pkg/errors"
)

var (

	// HpCPUExclusiveFlag should be used to specify that HP task will use cores assigned exclusively.
	HpCPUExclusiveFlag = conf.NewBoolFlag("hp_exclusive_cores", "Has high priority task exclusive cores", false)
	// BeCPUExclusiveFlag should be used to specify that BE task will use cores assigned exclusively.
	BeCPUExclusiveFlag = conf.NewBoolFlag("be_exclusive_cores", "Has best effort task exclusive cores", false)

	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)

	// HpSetsFlag should be used to set HP task cpuset.
	HpSetsFlag = conf.NewStringFlag("hp_sets", "HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	// BeSetsFlag should be used to set BE task cpuset.
	BeSetsFlag = conf.NewStringFlag("be_sets", "BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	// BeL1SetsFlag should be used to set BE task cpuset when one wants to provide L1 cache isolation.
	BeL1SetsFlag = conf.NewStringFlag("be_l1_sets", "BE for l1 aggressors cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
)

type defaultTopology struct {
	HpThreadIDs               isolation.IntSet
	SharingLLCButNotL1Threads isolation.IntSet
	SiblingThreadsToHpThreads topo.ThreadSet
	numaNode                  int
	isHpCPUExclusive          bool
	isBeCPUExclusive          bool
}

// NewIsolations returns HP anb factory of aggressors with applied isolation for BE tasks.
// TODO: needs update for different isolation per cpu
func NewIsolations() (hpIsolation, l1Isolation, llcIsolation isolation.Decorator) {
	if isManualPolicy() {
		manualTopology := NewManualTopology(HpSetsFlag.Value(), BeSetsFlag.Value(), BeL1SetsFlag.Value(), HpCPUExclusiveFlag.Value(), BeCPUExclusiveFlag.Value())
		llcIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.BeCPUs, PreferredNode: manualTopology.beNumaNodes[0]}
		l1Isolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.BeL1CPUs, PreferredNode: manualTopology.beL1NumaNodes[0]}
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.HpCPUs, PreferredNode: manualTopology.hpNumaNodes[0]}
	} else {
		defaultTopology, err := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value(), HpCPUExclusiveFlag.Value(), BeCPUExclusiveFlag.Value())
		errutil.Check(err)
		l1Isolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.SiblingThreadsToHpThreads.AvailableThreads().AsSlice(), PreferredNode: defaultTopology.numaNode}
		llcIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.SharingLLCButNotL1Threads.AsSlice(), PreferredNode: defaultTopology.numaNode}
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.HpThreadIDs.AsSlice(), PreferredNode: defaultTopology.numaNode}
	}
	return
}

func isManualPolicy() bool {
	return HpSetsFlag.Value() != "" && BeSetsFlag.Value() != ""
}

func newDefaultTopology(hpCPUCount, beCPUCount int, isHpCPUExclusive, isBeCPUExclusive bool) (defaultTopology, error) {
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

	topology.isHpCPUExclusive = isHpCPUExclusive
	topology.isBeCPUExclusive = isBeCPUExclusive

	return topology, nil
}

// ManualTopology represents flag-based cpuset topology for the experiment.
type ManualTopology struct {
	HpCPUs           []int
	hpNumaNodes      []int
	BeCPUs           []int
	beNumaNodes      []int
	BeL1CPUs         []int
	beL1NumaNodes    []int
	isHpCPUExclusive bool
	isBeCPUExclusive bool
}

// NewManualTopology prepares instance of ManualTopology based on flags that user passed.
func NewManualTopology(hpSets, beSets, beL1Sets string, isHpCPUExclusive, isBeCPUExclusive bool) ManualTopology {
	topology := ManualTopology{}
	topology.HpCPUs, topology.hpNumaNodes = parseSlices(hpSets)
	topology.BeCPUs, topology.beNumaNodes = parseSlices(beSets)
	topology.BeL1CPUs, topology.beL1NumaNodes = topology.BeCPUs, topology.beNumaNodes
	if beL1Sets != "" {
		topology.BeL1CPUs, topology.beL1NumaNodes = parseSlices(beL1Sets)
	}
	topology.isHpCPUExclusive = isHpCPUExclusive
	topology.isBeCPUExclusive = isBeCPUExclusive

	return topology
}

// parseSlices helper accepts raw string in format "1,2,3:5,3,1" and returns two slices of ints
func parseSlices(raw string) (CPUs, numaNodes []int) {
	// helper to parse slice of strings and return slice of ints
	parseInts := func(strings []string) (ints []int) {
		for _, s := range strings {
			i, err := strconv.Atoi(s)
			errutil.Check(err)
			ints = append(ints, i)
		}
		return
	}
	splits := strings.Split(raw, ":")
	s1Strings := strings.Split(splits[0], ",")
	s2Strings := strings.Split(splits[1], ",")
	CPUs = parseInts(s1Strings)
	numaNodes = parseInts(s2Strings)
	return
}
