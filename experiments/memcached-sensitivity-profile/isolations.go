package main

import (
	"strconv"
	"strings"

	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/isolation/topo"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
)

type defaultTopology struct {
	hpThreadIDs               isolation.IntSet
	sharingLLCButNotL1Threads isolation.IntSet
	siblingThreadsToHpThreads topo.ThreadSet
	numaNode                  int
	isHpCPUExclusive          bool
	isBeCPUExclusive          bool
}

func newDefaultTopology(hpCPUCouunt, beCPUCount int, isHpCPUExclusive, isBeCPUExclusive bool) defaultTopology {
	var topology defaultTopology
	var err error

	threadSet := sharedCacheThreads()
	topology.hpThreadIDs, err = threadSet.AvailableThreads().Take(hpCPUCouunt)
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
	isHpCPUExclusive bool
	isBeCPUExclusive bool
}

func newManualTopology(hpFlag, beFlag string, isHpCPUExclusive, isBeCPUExclusive bool) manualTopology {
	topology := manualTopology{}
	topology.hpCPUs, topology.hpNumaNodes = parseSlices(hpFlag)
	topology.beCPUs, topology.beNumaNodes = parseSlices(beFlag)
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
