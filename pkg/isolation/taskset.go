package isolation

import (
	"fmt"
)

// Taskset is wrapper for taskset linux tool to run process with CPU affinity.
type Taskset struct {
	CPUList IntSet
}

// Decorate command with taskset prefix.
func (ts Taskset) Decorate(command string) string {
	return fmt.Sprintf("taskset -c %s %s", ts.CPUList.AsRangeString(), command)
}
