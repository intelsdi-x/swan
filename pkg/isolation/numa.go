package isolation

import (
	"fmt"
	"strconv"
	"strings"
)

// Numa stores information about numactl configuration.
type Numa struct {
	isDisabledCPUAwareness bool
	shouldAllocOnLocalNode bool

	interleavePolicy   []int
	memoryNodes        []int
	cpusNodes          []int
	processCPUAffinity []int

	preferredMemoryNode int
}

// NewNuma is a constructor which returns Numa object.
// For further information please take a look into numactl manual.
func NewNuma(isDisabledCPUAwareness, shouldAllocOnLocalNode bool,
	interleavePolicy, memoryNodes, cpusNodes,
	processCPUAffinity []int, preferredMemoryNode int) Numa {
	return Numa{
		isDisabledCPUAwareness: isDisabledCPUAwareness,
		shouldAllocOnLocalNode: shouldAllocOnLocalNode,

		interleavePolicy:   interleavePolicy,
		memoryNodes:        memoryNodes,
		cpusNodes:          cpusNodes,
		processCPUAffinity: processCPUAffinity,

		preferredMemoryNode: preferredMemoryNode,
	}
}

// Decorate implements Decorator interface.
func (n *Numa) Decorate(command string) string {

	intsToStrings := func(ints []int) string {
		var strs []string
		for _, value := range ints {
			if value >= 0 {
				strs = append(strs, strconv.Itoa(value))
			}
		}
		return strings.Join(strs, ",")
	}

	var numaOptions []string
	if n.isDisabledCPUAwareness {
		numaOptions = append(numaOptions, "--all")
	}
	if n.shouldAllocOnLocalNode {
		numaOptions = append(numaOptions, "--localalloc")
	}
	if len(n.interleavePolicy) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--interleave=%s", intsToStrings(n.interleavePolicy)))
	}
	if len(n.memoryNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--membind=%s", intsToStrings(n.memoryNodes)))
	}
	if len(n.processCPUAffinity) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--physcpubind=%s", intsToStrings(n.processCPUAffinity)))
	}
	if len(n.cpusNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--cpunodebind=%s", intsToStrings(n.cpusNodes)))
	}
	if n.preferredMemoryNode > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--preferred=%d", n.preferredMemoryNode))
	}

	return fmt.Sprintf("numactl %s -- %s", strings.Join(numaOptions, " "), command)
}
