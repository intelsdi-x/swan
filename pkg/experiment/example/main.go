package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {
	fmt.Println("Nested Set() and Range()")

	var iteration int
	experiment.Set("one,two,three", func(context ...interface{}) {
		experiment.Range("1-3", func(context ...interface{}) {
			fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, context[0], context[1])
			iteration++
		})
	})

	fmt.Println("Should produce identical result as Permute()")

	iteration = 0
	experiment.Permute("one,two,three", "1-3", func(context ...interface{}) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, context[0], context[1])
		iteration++
	})
}
