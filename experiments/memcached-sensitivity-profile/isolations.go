package main

import (
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
)

var (
	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)

	// For manually provided isolation policy.
	hpSetsFlag = conf.NewStringFlag("hp_sets", "HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	beSetsFlag = conf.NewStringFlag("be_sets", "BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")

	hpCPUExclusiveFlag = conf.NewBoolFlag("hp_exclusive_cores", "Has high priority task exclusive cores", false)
	beCPUExclusiveFlag = conf.NewBoolFlag("be_exclusive_cores", "Has best effort task exclusive cores", false)
)

// sharedCacheIsolationPolicy TODO: describe intention of this policy.
func sensitivityProfileIsolationPolicy() (
	hpIsolation isolation.Isolation,
	siblingThreadsToHpThreadsIsolation isolation.Isolation,
	sharingLLCButNotL1Isolation isolation.Isolation) {

	threadSet := sharedCacheThreads()
	//hpThreadIDs, err := threadSet.AvailableThreads().Take(hpCPUCountFlag.Value())
	hpThreadIDs, err := threadSet.AvailableThreads().Take(2)
	check(err)

	// Allocate sibling threads of HP workload to create L1 cache contention
	threadSetOfHpThreads, err := topo.NewThreadSetFromIntSet(hpThreadIDs)
	check(err)
	siblingThreadsToHpThreads := getSiblingThreadsOfThreadSet(threadSetOfHpThreads)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(hpThreadIDs)
	//sharingLLCButNotL1Threads, err := remaining.Take(beCPUCountFlag.Value())
	sharingLLCButNotL1Threads, err := remaining.Take(2)
	check(err)

	// TODO(CD): Verify that it's safe to assume NUMA node 0 contains all.
	// memory banks (probably not).
	numaZero := isolation.NewIntSet(0)

	// Initialize Memcached Launcher with HP isolation.
	hpIsolation, err = cgroup.NewCPUSet(
		"hp",
		hpThreadIDs,
		numaZero,
		hpCPUExclusiveFlag.Value(),
		false)
	check(err)

	err = hpIsolation.Create()
	check(err)

	// Initialize BE L1 isolation.
	siblingThreadsToHpThreadsIsolation, err = cgroup.NewCPUSet(
		"be-l1",
		siblingThreadsToHpThreads.AvailableThreads(),
		numaZero,
		beCPUExclusiveFlag.Value(),
		false)
	check(err)

	err = siblingThreadsToHpThreadsIsolation.Create()
	check(err)

	// Initialize BE LLC isolation.
	sharingLLCButNotL1Isolation, err = cgroup.NewCPUSet(
		"be",
		sharingLLCButNotL1Threads,
		numaZero,
		beCPUExclusiveFlag.Value(),
		false)
	check(err)

	err = sharingLLCButNotL1Isolation.Create()
	check(err)

	logrus.Infof("Sensitivity Profile Isolation:\n"+
		"High Priority Job CpuThreads: %v\n"+
		"L1 Cache Aggressor CpuThreads: %v\n"+
		"L3 Cache  Aggressor CpuThreads: %v\n",
		hpThreadIDs, siblingThreadsToHpThreads.AvailableThreads(), sharingLLCButNotL1Threads)

	return hpIsolation, siblingThreadsToHpThreadsIsolation, sharingLLCButNotL1Isolation
}

// parseSlices helper accepts raw string in format "1,2,3:5,3,1" and returns two slices of ints
func parseSlices(raw string) (s1, s2 []int) {
	// helper to parse slice of strings and return slice of ints
	parseInts := func(strings []string) (ints []int) {
		for _, s := range strings {
			i, err := strconv.Atoi(s)
			check(err)
			ints = append(ints, i)
		}
		return
	}
	splits := strings.Split(raw, ":")
	s1Strings := strings.Split(splits[0], ",")
	s2Strings := strings.Split(splits[1], ",")
	s1 = parseInts(s1Strings)
	s2 = parseInts(s2Strings)
	return
}

// manualPolicy helper to create HP and BE isolations based on manually provided flags (--hp_sets and --be_sets).
func manualPolicy() (hpIsolation, beIsolation isolation.Isolation) {

	// TODO: validation of input data: cpus and numa node overlap and no empty
	hpCPUs, hpNUMAs := parseSlices(hpSetsFlag.Value())
	beCPUs, beNUMAs := parseSlices(beSetsFlag.Value())

	logrus.Debugf("HP: CPUs=%v NUMAs=%v", hpCPUs, hpNUMAs)
	logrus.Debugf("BE: CPUs=%v NUMAs=%v", beCPUs, beNUMAs)

	var err error
	hpIsolation, err = cgroup.NewCPUSet("hp", isolation.NewIntSet(hpCPUs...), isolation.NewIntSet(hpNUMAs...), hpCPUExclusiveFlag.Value(), false)
	check(err)
	beIsolation, err = cgroup.NewCPUSet("be", isolation.NewIntSet(beCPUs...), isolation.NewIntSet(beNUMAs...), beCPUExclusiveFlag.Value(), false)
	check(err)

	err = hpIsolation.Create()
	check(err)
	err = beIsolation.Create()
	check(err)
	return
}
