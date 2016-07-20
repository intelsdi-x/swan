package isolation

import (
	"fmt"
	"strconv"
	"strings"
)

// Numa stores information about numactl configuration.
type Numa struct {
	isAll        bool
	isLocalalloc bool

	interleave  []int
	membind     []int
	cpunodebind []int
	physcpubind []int

	preffered int
}

// NewNuma is a constructor which returns Numa object.
// For further information please take a look into numactl manual.
func NewNuma(all, localalloc bool, interleave, membind, cpunodebind, physcpubind []int, preffered int) Numa {
	return Numa{
		isAll:        all,
		isLocalalloc: localalloc,

		interleave:  interleave,
		membind:     membind,
		cpunodebind: cpunodebind,
		physcpubind: physcpubind,

		preffered: preffered,
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
	if n.isAll {
		numaOptions = append(numaOptions, "-a")
	}
	if n.isLocalalloc {
		numaOptions = append(numaOptions, "-l")
	}
	if len(n.interleave) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("-i %s", intsToStrings(n.interleave)))
	}
	if len(n.membind) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("-m %s", intsToStrings(n.membind)))
	}
	if len(n.physcpubind) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("-C %s", intsToStrings(n.physcpubind)))
	}
	if len(n.cpunodebind) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("-N %s", intsToStrings(n.cpunodebind)))
	}
	if n.preffered > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--preffered=%d", n.preffered))
	}

	return fmt.Sprintf("numactl %s -- %s", strings.Join(numaOptions, " "), command)
}
