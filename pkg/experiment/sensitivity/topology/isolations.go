package topology

import (
	"strconv"
	"strings"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

var (

	// Should CPUs be used exclusively?
	hpCPUExclusiveFlag = conf.NewBoolFlag("hp_exclusive_cores", "Has high priority task exclusive cores", false)
	beCPUExclusiveFlag = conf.NewBoolFlag("be_exclusive_cores", "Has best effort task exclusive cores", false)

	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)

	// For manually provided isolation policy.
	hpSetsFlag   = conf.NewStringFlag("hp_sets", "HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	beSetsFlag   = conf.NewStringFlag("be_sets", "BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	beL1SetsFlag = conf.NewStringFlag("be_l1_sets", "BE for l1 aggressors cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
)

type defaultTopology struct {
	hpThreadIDs               isolation.IntSet
	sharingLLCButNotL1Threads isolation.IntSet
	siblingThreadsToHpThreads topo.ThreadSet
	numaNode                  int
	isHpCPUExclusive          bool
	isBeCPUExclusive          bool
}

// NewIsolations returns HP anb factory of aggressors with applied isolation for BE tasks.
// TODO: needs update for different isolation per cpu
func NewIsolations() (hpIsolation, l1Isolation, llcIsolation isolation.Decorator) {
	if isManualPolicy() {
		manualTopology := newManualTopology(hpSetsFlag.Value(), beSetsFlag.Value(), beL1SetsFlag.Value(), hpCPUExclusiveFlag.Value(), beCPUExclusiveFlag.Value())
		llcIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.beCPUs, PreferredNode: manualTopology.beNumaNodes[0]}
		l1Isolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.beL1CPUs, PreferredNode: manualTopology.beL1NumaNodes[0]}
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.hpCPUs, PreferredNode: manualTopology.hpNumaNodes[0]}
	} else {
		defaultTopology := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value(), hpCPUExclusiveFlag.Value(), beCPUExclusiveFlag.Value())
		l1Isolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.siblingThreadsToHpThreads.AvailableThreads().AsSlice(), PreferredNode: defaultTopology.numaNode}
		llcIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.sharingLLCButNotL1Threads.AsSlice(), PreferredNode: defaultTopology.numaNode}
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.hpThreadIDs.AsSlice(), PreferredNode: defaultTopology.numaNode}
	}
	return
}

func isManualPolicy() bool {
	return hpSetsFlag.Value() != "" && beSetsFlag.Value() != ""
}

func newDefaultTopology(hpCPUCount, beCPUCount int, isHpCPUExclusive, isBeCPUExclusive bool) defaultTopology {
	var topology defaultTopology
	var err error

	threadSet := sharedCacheThreads()
	topology.hpThreadIDs, err = threadSet.AvailableThreads().Take(hpCPUCount)
	errutil.Check(err)

	// Allocate sibling threads of HP workload to create L1 cache contention
	threadSetOfHpThreads, err := topo.NewThreadSetFromIntSet(topology.hpThreadIDs)
	errutil.Check(err)
	topology.siblingThreadsToHpThreads = getSiblingThreadsOfThreadSet(threadSetOfHpThreads)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(topology.hpThreadIDs)
	topology.sharingLLCButNotL1Threads, err = remaining.Take(beCPUCount)
	errutil.Check(err)

	topology.isHpCPUExclusive = isHpCPUExclusive
	topology.isBeCPUExclusive = isBeCPUExclusive

	return topology
}

type manualTopology struct {
	hpCPUs           []int
	hpNumaNodes      []int
	beCPUs           []int
	beNumaNodes      []int
	beL1CPUs         []int
	beL1NumaNodes    []int
	isHpCPUExclusive bool
	isBeCPUExclusive bool
}

func newManualTopology(hpSets, beSets, beL1Sets string, isHpCPUExclusive, isBeCPUExclusive bool) manualTopology {
	topology := manualTopology{}
	topology.hpCPUs, topology.hpNumaNodes = parseSlices(hpSets)
	topology.beCPUs, topology.beNumaNodes = parseSlices(beSets)
	topology.beL1CPUs, topology.beL1NumaNodes = topology.beCPUs, topology.beNumaNodes
	if beL1Sets != "" {
		topology.beL1CPUs, topology.beL1NumaNodes = parseSlices(beL1Sets)
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
