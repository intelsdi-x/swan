package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {

	var iteration int
	repetition := func(a string, b int) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
		iteration++
	}

	fmt.Println("Nested Iterate() calls")
	experiment.Iterate(experiment.Arg{"some random set", "one,two,three"}, func() {
		experiment.Iterate(experiment.Arg{"equally random interval", "1-3"}, repetition)
	})

	fmt.Println("Should produce identical result as a Permutate() call")
	iteration = 0
	experiment.Permutate(experiment.Arg{"some random set", "one,two,three"}, experiment.Arg{"equally random interval", "1-3"}, repetition)

	fmt.Println("Dry run should ouput nothing")
	experiment.Permutate(experiment.Arg{"some random set", "one,two,three"}, experiment.Arg{"equally random interval", "1-3"}, experiment.DryRun)
	fmt.Printf("Number of iterations that would have been executed: %d", experiment.GetNumberOfIterations())
}
