package isolation

import (
	"testing"
	"fmt"
)

func Testcpu(t *testing.T) {
	cpu := cpuShares{name: "Hello"}
	
	fmt.Printf(cpu.name)

}
