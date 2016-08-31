package main

import (
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/isolation/cgroup"
	"github.com/intelsdi-x/athena/pkg/isolation/topo"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
)

func createNumactlIsolation(topology defaultTopology) (
	hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation isolation.Decorator) {
	hpIsolation = isolation.NewNumactl(false, false, []int{}, []int{}, []int{}, topology.hpThreadIDs.AsSlice(), topology.numaNode)
	siblingThreadsToHpThreadsIsolation = isolation.NewNumactl(
		false,
		false,
		[]int{},
		[]int{},
		[]int{},
		topology.siblingThreadsToHpThreads.AvailableThreads().AsSlice(),
		topology.numaNode)
	sharingLLCButNotL1Isolation = isolation.NewNumactl(false, false, []int{}, []int{}, []int{}, topology.sharingLLCButNotL1Threads.AsSlice(), topology.numaNode)

	return
}

func createCPUsetIsolation(topology defaultTopology) (
	hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation isolation.Decorator) {

	// TODO(CD): Verify that it's safe to assume NUMA node 0 contains all.
	// memory banks (probably not).
	numaZero := isolation.NewIntSet(topology.numaNode)

	// Initialize Memcached Launcher with HP isolation.
	hpIsolation, err := cgroup.NewCPUSet(
		"hp",
		topology.hpThreadIDs,
		numaZero,
		topology.isHpCPUExclusive,
		false)
	errutil.Check(err)
	createIsolation(hpIsolation)

	// Initialize BE L1 isolation.
	siblingThreadsToHpThreadsIsolation, err = cgroup.NewCPUSet(
		"be-l1",
		topology.siblingThreadsToHpThreads.AvailableThreads(),
		numaZero,
		topology.isBeCPUExclusive,
		false)
	errutil.Check(err)
	createIsolation(siblingThreadsToHpThreadsIsolation)

	// Initialize BE LLC isolation.
	sharingLLCButNotL1Isolation, err = cgroup.NewCPUSet(
		"be",
		topology.sharingLLCButNotL1Threads,
		numaZero,
		topology.isBeCPUExclusive,
		false)
	errutil.Check(err)
	createIsolation(sharingLLCButNotL1Isolation)

	logrus.Infof("Sensitivity Profile Isolation:\n"+
		"High Priority Job CpuThreads: %v\n"+
		"L1 Cache Aggressor CpuThreads: %v\n"+
		"L3 Cache  Aggressor CpuThreads: %v\n",
		topology.hpThreadIDs, topology.siblingThreadsToHpThreads.AvailableThreads(), topology.sharingLLCButNotL1Threads)

	return
}

func createDefaultIsolation(topology defaultTopology, isK8s bool) (
	hpIsolation isolation.Decorator,
	siblingThreadsToHpThreadsIsolation isolation.Decorator,
	sharingLLCButNotL1Isolation isolation.Decorator) {
	if isK8s {
		hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation = createNumactlIsolation(topology)
	} else {
		hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation = createCPUsetIsolation(topology)
	}

	return hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation
}

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

func createManualIsolation(topology manualTopology, isK8s bool) (hpIsolation, beIsolation isolation.Decorator) {
	logrus.Debugf("HP: CPUs=%v NUMAs=%v", topology.hpCPUs, topology.hpNumaNodes)
	logrus.Debugf("BE: CPUs=%v NUMAs=%v", topology.beCPUs, topology.beNumaNodes)

	if isK8s {
		hpIsolation = isolation.NewNumactl(false, false, []int{}, []int{}, []int{}, topology.hpCPUs, topology.hpNumaNodes[0])
		beIsolation = isolation.NewNumactl(false, false, []int{}, []int{}, []int{}, topology.beCPUs, topology.beNumaNodes[0])
	} else {
		var err error
		hpIsolation, err = cgroup.NewCPUSet("hp", isolation.NewIntSet(topology.hpCPUs...), isolation.NewIntSet(topology.hpNumaNodes...), topology.isHpCPUExclusive, false)
		errutil.Check(err)
		beIsolation, err = cgroup.NewCPUSet("be", isolation.NewIntSet(topology.beCPUs...), isolation.NewIntSet(topology.beNumaNodes...), topology.isBeCPUExclusive, false)
		errutil.Check(err)
	}
	createIsolation(hpIsolation)
	createIsolation(beIsolation)

	return
}

func createIsolation(decorator isolation.Decorator) {
	isolator, ok := decorator.(isolation.Isolation)
	if ok {
		isolator.Create()
	}
}

func cleanIsolation(decorator isolation.Decorator) {
	isolator, ok := decorator.(isolation.Isolation)
	if ok {
		err := isolator.Clean()
		errutil.Check(err)
	}
}
