package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {
	fmt.Println("Nested Set() and Range()")

	var iteration int
	experiment.Set("one,two,three", func(setValue interface{}) {
		experiment.Range("1-3", func(rangeValue interface{}) {
			fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, setValue, rangeValue)
			iteration++
		})
	})

	fmt.Println("Should produce identical result as Permute()")

	iteration = 0
	experiment.Permute([]string{"one,two,three", "1-3"}, func(rangeValue interface{}) {
		fmt.Printf("iteration: %d\tset: \trange: %v\n", iteration, rangeValue)
		iteration++
	})
}
