package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
    "strings"
    "fmt"

	"github.com/pkg/errors"
)

const (
    // ALL - Unset default cpuset awareness.
    ALL = 1 << iota
    // INTERLEAVE - Set a memory interleave policy. Memory will be allocated using round robin on nodes.
    INTERLEAVE
    // MEMBIND - Only allocate memory from nodes.
    MEMBIND
    // CPUNODEBIND - Only execute command on the CPUs of nodes.
    CPUNODEBIND
    // PHYSCPUBIND - Only execute process on cpus.
    PHYSCPUBIND
    // LOCALALLOC - Always allocate on the current node.
    LOCALALLOC
    // PREFFERED - Preferably allocate memory on node, but if memory cannot be allocated there fall back to other nodes.
    PREFFERED
)

// Numa defines input data
type Numa struct {
	numaOpts map[int]string
}
 

// NewNuma is 
func NewNuma(all, localalloc bool, interleave, membind, cpunodebind, physcpubind []int, preffered int) Numa {

    intsToStrings := func(ints []int) []string {
        strs := make([]string, len(ints))
        for idx, value := range ints {
            strs[idx] := strconv.Itoa(value)
        }
        return strs
    }

    numaOpts := make(map[int]string)
    if interleave == true {
        numaOpts[ALL] = "-a"
    }
    if localalloc == true {
        numaOpts[LOCALALLOC] = "-l"
    }

    if len(interleave) > 0 {
        numaOpts[INTERLEAVE] = fmt.Sprintf("-i %s", strings.Join(intsToStrings(interleave), ","))
    }

    if len(membind) > 0 {
        numaOpts[MEMBIND] = fmt.Sprintf("-m %s", strings.Join(intsToStrings(membind), ","))
    }

    if len(cpunodebind) > 0 {
        numaOpts[CPUNODEBIND] = fmt.Sprintf("-N %s", strings.Join(intsToStrings(cpunodebind), ",")
    }

    if len(physcpubind) > 0 {
        numaOpts[PHYSCPUBIND] = fmt.Sprintf("-C %s", strings.Join(intsToStrings(physcpubind), ","))
    }

    numaOpts[PREFFERED] = preffered

    return Numa{
        numaOpts: numaOpts,
    }

}

// Decorate implements Decorator interface.
func (n *Numa) Decorate(command string) string {
    var numaOptions []string
    for _, numaOption := range n.numaOpts {
        numaOptions = append(numaOptions, numaOption)
    }
	return fmt.Sprintf("numactl %s %s", strings.Join(numaOptions), command)
}
