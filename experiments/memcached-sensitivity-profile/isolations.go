package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
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

// sharedCacheIsolationPolicy TODO: describe intention of this policy
func sharedCacheIsolationPolicy() (hpIsolation, beIsolation isolation.Isolation) {

	threadSet := sharedCacheThreads()
	hpThreadIDs, err := threadSet.AvailableThreads().Take(hpCPUCountFlag.Value())
	check(err)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(hpThreadIDs)
	beThreadIDs, err := remaining.Take(beCPUCountFlag.Value())
	check(err)

	// TODO(CD): Verify that it's safe to assume NUMA node 0 contains all
	// memory banks (probably not).
	numaZero := isolation.NewIntSet(0)

	// Initialize Memcached Launcher with HP isolation.
	hpIsolation, err = cgroup.NewCPUSet("hp", hpThreadIDs, numaZero, hpCPUExclusiveFlag.Value(), false)
	check(err)

	err = hpIsolation.Create()
	check(err)

	// Initialize BE isolation.
	beIsolation, err = cgroup.NewCPUSet("be", beThreadIDs, numaZero, beCPUExclusiveFlag.Value(), false)
	check(err)

	err = beIsolation.Create()
	check(err)

	return
}

// manualPolicy helper to create HP and BE isolations based on manually provided flags.
func manualPolicy() (hpIsolation, beIsolation isolation.Isolation) {

	parse := func(raw string) (CPUs, NUMAs []int) {
		parseInts := func(strings []string) (ints []int) {
			for _, s := range strings {
				i, err := strconv.Atoi(s)
				check(err)
				ints = append(ints, i)
			}
			return
		}
		CPUStrings := strings.Split(strings.Split(raw, ":")[0], ",")
		NUMAStrings := strings.Split(strings.Split(raw, ":")[1], ",")
		CPUs = parseInts(CPUStrings)
		NUMAs = parseInts(NUMAStrings)
		return
	}

	hpCPUs, hpNUMAs := parse(hpSetsFlag.Value())
	beCPUs, beNUMAs := parse(beSetsFlag.Value())

	logrus.Debugf("HP: CPUs=%v NUMAs=%v", hpCPUs, hpNUMAs)
	logrus.Debugf("BE: CPUs=%v NUMAs=%v", beCPUs, beNUMAs)
	os.Exit(1)

	var err error
	hpIsolation, err = cgroup.NewCPUSet("hp", isolation.NewIntSet(hpCPUs...), isolation.NewIntSet(hpNUMAs...), hpCPUExclusiveFlag.Value(), false)
	check(err)
	beIsolation, err = cgroup.NewCPUSet("be", isolation.NewIntSet(beCPUs...), isolation.NewIntSet(beNUMAs...), beCPUExclusiveFlag.Value(), false)
	check(err)
	return
}
