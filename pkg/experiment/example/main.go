package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {
	// We need a function that will output something to the screen. It will calculate total number of iterations too.
	var iteration int
	repetition := func(a string, b float64) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
		iteration++
	}

	// Nested Iterate() calls
	fmt.Println("Nested Iterate() calls")
	experiment.Iterate(experiment.Arg{"some random set", experiment.Set{"one", "two", "three"}}, func() {
		experiment.Iterate(experiment.Arg{"equally random interval", &experiment.Range{From: 0, To: 3, Step: 1}}, repetition)
	})

	// Single Permutate() call - it is equivalent od the nested Iterate() implementation.
	fmt.Println("Should produce identical result as a Permutate() call")
	iteration = 0
	experiment.Permutate(experiment.Arg{"some random set", experiment.Set{"one", "two", "three"}}, experiment.Arg{"equally random interval", &experiment.Range{From: 0, To: 3, Step: 1}}, repetition)

	// Dry run example - it allows to calculate number of iterations that Permutate() is to generate.
	fmt.Println("Dry run should ouput nothing")
	experiment.Permutate(experiment.Arg{"some random set", experiment.Set{"one", "two", "three"}}, experiment.Arg{"equally random interval", &experiment.Range{From: 0, To: 3, Step: 0.001}}, experiment.DryRun)
	fmt.Printf("Number of iterations that would have been executed: %d", experiment.GetNumberOfIterations())
}
