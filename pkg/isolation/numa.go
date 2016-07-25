package isolation

import (
	"fmt"
	"strconv"
	"strings"
)

// Numactl stores information about numactl configuration.
// All parameters are described in numactl manual.
// http://linux.die.net/man/8/numactl
type Numactl struct {
	isAll        bool
	isLocalalloc bool

	interleaveNodes  []int
	membindNodes     []int
	cpunodebindNodes []int
	physcpubindCPUS  []int

	preferredNode int
}

// NewNumactl is a constructor which returns Numa object.
func NewNumactl(isAll, isLocalalloc bool, interleaveNodes, membindNodes, cpunodebindNodes, physcpubindCPUS []int, preferredNode int) Numactl {
	return Numactl {
		isAll:        isAll,
		isLocalalloc: isLocalalloc,

		interleaveNodes:  interleaveNodes,
		membindNodes:     membindNodes,
		cpunodebindNodes: cpunodebindNodes,
		physcpubindCPUS:  physcpubindCPUS,

		preferredNode: preferredNode,
	}
}

// Decorate implements Decorator interface.
func (n *Numactl) Decorate(command string) string {

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
	if n.isAll {
		numaOptions = append(numaOptions, "--all")
	}
	if n.isLocalalloc {
		numaOptions = append(numaOptions, "--localalloc")
	}
	if len(n.interleaveNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--interleave=%s", intsToStrings(n.interleaveNodes)))
	}
	if len(n.membindNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--membind=%s", intsToStrings(n.membindNodes)))
	}
	if len(n.physcpubindCPUS) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--physcpubind=%s", intsToStrings(n.physcpubindCPUS)))
	}
	if len(n.cpunodebindNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--cpunodebind=%s", intsToStrings(n.cpunodebindNodes)))
	}
	if n.preferredNode > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--preferred=%d", n.preferredNode))
	}

	return fmt.Sprintf("numactl %s -- %s", strings.Join(numaOptions, " "), command)
}
