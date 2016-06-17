package topo

import (
	"fmt"
	"os/exec"
	"strings"
)

// Discover CPU and basic NUMA topology.
func Discover() (ThreadSet, error) {
	out, err := exec.Command("lscpu", "-p").Output()
	if err != nil {
		return nil, err
	}
	return ReadTopology(out)
}

// ReadTopology attempts to create a ThreadSet that corresponds to the
// supplied output from `lscpu -p`.
func ReadTopology(lscpuOutput []byte) (ThreadSet, error) {
	threadSet := NewThreadSet()

	out := strings.TrimSpace(string(lscpuOutput))
	lines := strings.Split(out, "\n")

	// lscpu -p output looks like:
	//
	// # comments
	// # comments
	// cpu,core,socket,node,,l1d,l1i,l2,l3
	// cpu,core,socket,node,,l1d,l1i,l2,l3
	// ...
	for _, line := range lines {
		// Skip informational header lines.
		if strings.HasPrefix(line, "#") {
			continue
		}

		var cpu, core, socket int
		n, err := fmt.Sscanf(line, "%d,%d,%d", &cpu, &core, &socket)
		if n != 3 {
			return nil, fmt.Errorf("Expected to read 3 values but got %d", n)
		}
		if err != nil {
			return nil, err
		}

		// Construct a new thread and append it to the "set".
		threadSet = append(threadSet, NewThread(cpu, core, socket))
	}

	return threadSet, nil
}
