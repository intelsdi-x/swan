package isolation

import (
	"fmt"
	"github.com/pivotal-golang/bytefmt"
	"os/exec"
	"strconv"
	"strings"
)

// CPUInfo defines data needed for CPU topology.
type CPUInfo struct {
	Sockets        int
	PhysicalCores  int
	ThreadsPerCore int
	CacheL1i       int
	CacheL1d       int
	CacheL2        int
	CacheL3        int

	// Maps physical sockets to physical cores
	SocketCores map[int]IntSet

	// Maps physical cores to logical CPU ids
	CoreCpus map[int]IntSet
}

// NewCPUInfo instance creation.
func NewCPUInfo(processors int, cores int, threads int, l1i int, l1d int, l2 int, l3 int) *CPUInfo {
	return &CPUInfo{Sockets: processors, PhysicalCores: cores, ThreadsPerCore: threads, CacheL1i: l1i, CacheL1d: l1d, CacheL2: l2, CacheL3: l3}
}

// Discover CPU topology and caches sizes. We use lscpu for the time being until we can make this code more portable by reading directly from HW.
func (cputopo *CPUInfo) Discover() error {
	parseMap := map[string]*int{
		"Socket(s)":          &cputopo.Sockets,
		"Core(s) per socket": &cputopo.PhysicalCores,
		"Thread(s) per core": &cputopo.ThreadsPerCore,
		"L1i cache":          &cputopo.CacheL1i,
		"L1d cache":          &cputopo.CacheL1d,
		"L2 cache":           &cputopo.CacheL2,
		"L3 cache":           &cputopo.CacheL3,
	}

	out, err := exec.Command("lscpu").Output()
	if err != nil {
		return err
	}

	outstring := strings.TrimSpace(string(out))
	lines := strings.Split(outstring, "\n")

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		ptr, ok := parseMap[key]
		if ok {
			bytes, err := bytefmt.ToBytes(value)
			if err == nil {
				*ptr = int(bytes)
				continue

			}

			t, err := strconv.Atoi(value)
			if err != nil {
				return err
			}

			*ptr = t
		}
	}

	// Build core topology mappings
	out, err = exec.Command("lscpu", "-p").Output()
	if err != nil {
		return err
	}

	outstring = strings.TrimSpace(string(out))
	lines = strings.Split(outstring, "\n")

	socketCores := make(map[int]IntSet)
	coreCpus := make(map[int]IntSet)

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		// cpu,core,socket,node,,l1d,l1i,l2,l3
		var cpu, core, socket int
		n, err := fmt.Sscanf(line, "%d,%d,%d", &cpu, &core, &socket)
		if n != 3 {
			return fmt.Errorf("Expected to read 3 values but got %d", n)
		}
		if err != nil {
			return err
		}
		// Check whether we know about this socket yet
		if _, found := socketCores[socket]; !found {
			socketCores[socket] = NewIntSet()
		}
		// Save the socket -> core mapping
		socketCores[socket].Add(core)

		// Check whether we know about this core yet
		if _, found := coreCpus[core]; !found {
			coreCpus[core] = NewIntSet()
		}
		// Save the core -> cpu mapping
		coreCpus[core].Add(cpu)
	}

	cputopo.SocketCores = socketCores
	cputopo.CoreCpus = coreCpus

	return nil
}
