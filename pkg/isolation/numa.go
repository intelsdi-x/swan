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
	IsAll        bool
	IsLocalalloc bool

	InterleaveNodes  []int
	MembindNodes     []int
	CPUnodebindNodes []int
	PhyscpubindCPUs  []int

	PreferredNode int
}

// Decorate implements Decorator interface.
func (n Numactl) Decorate(command string) string {

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
	if n.IsAll {
		numaOptions = append(numaOptions, "--all")
	}
	if n.IsLocalalloc {
		numaOptions = append(numaOptions, "--localalloc")
	}
	if len(n.InterleaveNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--interleave=%s", intsToStrings(n.InterleaveNodes)))
	}
	if len(n.MembindNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--membind=%s", intsToStrings(n.MembindNodes)))
	}
	if len(n.PhyscpubindCPUs) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--physcpubind=%s", intsToStrings(n.PhyscpubindCPUs)))
	}
	if len(n.CPUnodebindNodes) > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--cpunodebind=%s", intsToStrings(n.CPUnodebindNodes)))
	}
	if n.PreferredNode > 0 {
		numaOptions = append(numaOptions, fmt.Sprintf("--preferred=%d", n.PreferredNode))
	}

	return fmt.Sprintf("numactl %s -- %s", strings.Join(numaOptions, " "), command)
}
