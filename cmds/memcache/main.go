// dummy package to make sure that we depdencies are bundled correclty
package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/dummy"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
)

// dummy to execrsise the workloads
func main() {
	fmt.Printf("dummy = %+v\n", dummy.Dummy{})
	fmt.Printf("remote = %+v\n", executor.NewLocal())
	var l workloads.Launcher
	fmt.Printf("workloads = %+v\n", l)
}
