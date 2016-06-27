package cgroup

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

// CPUSet represents a cgroup in the cpuset hierarchy.
type CPUSet interface {
	isolation.Isolation

	// Cgroup returns the underlying cgroup for this CPUSet.
	Cgroup() Cgroup

	// Cpus returns the set of cpus allocated to this CPUSet.
	Cpus() isolation.IntSet

	// Cpus returns the set of memory nodes allocated to this CPUSet.
	Mems() isolation.IntSet

	// CPUExclusive returns true if this CPUSet's cpus are allocated
	// exclusively..
	CPUExclusive() bool

	// MemExclusive returns true if this CPUSet's memory nodes are allocated
	// exclusively..
	MemExclusive() bool
}

const (
	// CPUSetMems is the name of the memory node attribute for a cpuset.
	CPUSetMems = "cpuset.mems"

	// CPUSetCpus is the name of the cpus attribute for a cpuset.
	CPUSetCpus = "cpuset.cpus"

	// CPUSetCPUExclusive is the name of the exclusive cpu attribute
	// for a cpuset.
	CPUSetCPUExclusive = "cpuset.cpu_exclusive"

	// CPUSetMemExclusive is the name of the exclusive memory node attribute
	// for a cpuset.
	CPUSetMemExclusive = "cpuset.mem_exclusive"
)

// CPUSet describes a cgroup cpuset with core ids and numa (memory) nodes.
type cpuset struct {
	cgroup       Cgroup
	cpus         isolation.IntSet
	mems         isolation.IntSet
	cpuExclusive bool
	memExclusive bool
}

// NewCPUSet creates a new CPUSet with the default (local) executor
// and default timeout.
func NewCPUSet(path string, cpus, mems isolation.IntSet, cpuExclusive, memExclusive bool) (CPUSet, error) {
	return NewCPUSetWithExecutor(path, cpus, mems, cpuExclusive, memExclusive, executor.NewLocal(), DefaultCommandTimeout)
}

// NewCPUSetWithExecutor creates a new CPUSet with the supplied executor
// and timeout.
func NewCPUSetWithExecutor(path string,
	cpus isolation.IntSet,
	mems isolation.IntSet,
	cpuExclusive bool,
	memExclusive bool,
	executor executor.Executor,
	timeout time.Duration) (CPUSet, error) {
	// Construct underlying cgroup.
	cg, err := NewCgroupWithExecutor([]string{CPUSetController}, path, executor, timeout)
	if err != nil {
		return nil, err
	}
	if len(cpus) == 0 {
		return nil, fmt.Errorf("Empty set of cpus provided")
	}
	if len(mems) == 0 {
		return nil, fmt.Errorf("Empty set of memory nodes provided")
	}

	cs := &cpuset{
		cgroup:       cg,
		cpus:         cpus,
		mems:         mems,
		cpuExclusive: cpuExclusive,
		memExclusive: memExclusive,
	}
	return cs, nil
}

func (cs *cpuset) Cgroup() Cgroup {
	return cs.cgroup
}

func (cs *cpuset) Cpus() isolation.IntSet {
	return cs.cpus
}

func (cs *cpuset) Mems() isolation.IntSet {
	return cs.mems
}

func (cs *cpuset) CPUExclusive() bool {
	return cs.cpuExclusive
}

func (cs *cpuset) MemExclusive() bool {
	return cs.memExclusive
}

// Decorate implements Decorator interface
func (cs *cpuset) Decorate(command string) string {
	return cs.cgroup.Decorate(command)
}

// Clean removes the underlying cgroup.
func (cs *cpuset) Clean() error {
	return cs.cgroup.Clean()
}

// Create instantiates the underlying cgroup and sets up the necessary
// attributes.
func (cs *cpuset) Create() error {
	// Create the cgroup.
	err := cs.cgroup.Create()
	if err != nil {
		cs.Clean()
		return err
	}

	// When setting cpuset.cpus or cpuset.mems, overwrite everything
	// except root. When starting experiment it's neccessary to be
	// sure that there are proper values and not leftovers from
	// earlier crash.
	//
	// When setting cpuset.cpu_exclusive or cpuset.mem_exclusive, the
	// attribute must first be set for all cgroup ancestors, starting with
	// the root of the hierarchy. If this is not done first, setting the
	// attribute will fail! These values default to "0" (off) for all
	// non-root cgroups.
	// Note also that cpu_exclusive cannot be set if any sibling cpuset
	// overlaps with current cgroup.

	for _, a := range cs.cgroup.Ancestors() {
		if a.IsRoot() {
			continue
		}
		// Setup parents.
		err = cs.setupCgroup(a)
		if err != nil {
			cs.Clean()
			return err
		}
	}

	err = cs.setupCgroup(cs.cgroup)
	if err != nil {
		cs.Clean()
		return err
	}

	return nil
}

// Isolate moves the process with id PID into the underlying cgroup.
// The cgroup must exist first
func (cs *cpuset) Isolate(PID int) error {
	return cs.cgroup.Isolate(PID)
}

func (cs *cpuset) setupCgroup(c Cgroup) error {
	err := c.Set(CPUSetCpus, cs.cpus.AsRangeString())
	if err != nil {
		return err
	}

	err = c.Set(CPUSetMems, cs.mems.AsRangeString())
	if err != nil {
		return err
	}

	// Set cpu exclusivity bit if necessary.
	if cs.cpuExclusive {
		err = c.SetAndCheck(CPUSetCPUExclusive, "1")
		if err != nil {
			return err
		}
	}

	// Set memory exclusivity bit if necessary.
	if cs.memExclusive {
		err = c.SetAndCheck(CPUSetMemExclusive, "1")
		if err != nil {
			return err
		}
	}

	return nil
}
