package experiment

import (
	"fmt"
	"reflect"
)

// Iterator is an interface that struct needs to implement in order to be part of Arg to allow Iterate() to consume it
type Iterator interface {
	Iterate(interface{})
}

var experimentContext []interface{}

var numberOfIterations int

// Arg represents named iteration that Permutate() and Iterate() can consume
type Arg struct {
	Name string
	Spec Iterator
}

// Permutate accepts variable number of arguments. Allows to specify multiple instances of Arg of various type of Spec. The last one needs to be a function that is to be executed.
func Permutate(specs ...interface{}) {
	if len(specs) > 2 {
		recursive := func() {
			Permutate(specs[1:]...)
		}
		Iterate(specs[0].(Arg), recursive)
	} else {

		Iterate(specs[0].(Arg), specs[1])
	}
}

// Iterate is a public interface of the experiment. It can recognize type of iteration and runs approptiate code. Last argument needs to be a function that is to be executed.
func Iterate(spec Arg, runnable interface{}) {
	fmt.Printf("Executing step %s with arguments %s\n", spec.Name, spec.Spec)
	spec.Spec.Iterate(runnable)
}

// Call handles calling a function passed to Iterate(). It is capable of calling a function with no arguments or with number of arguments equal to number of Arg instances parsed.
func Call(runnable, localContext interface{}) {
	experimentContext = append(experimentContext, localContext)
	defer func() { experimentContext = experimentContext[:len(experimentContext)-1] }()
	function := reflect.ValueOf(runnable)
	if function.Type().NumIn() == len(experimentContext) {
		// number of runnable arguments should be equal to current context size (including localContext)...
		var args []reflect.Value
		for _, v := range experimentContext {
			args = append(args, reflect.ValueOf(v))
		}
		function.Call(args)
	} else {
		// ... or it should have no arguments at all
		function.Call(nil)
	}
}

// DryRun is a helper function that allows to estimate number of iterations that Permutate() generates
func DryRun() {
	numberOfIterations++
}

// GetNumberOfIterations retrieves number of dry run iterations
func GetNumberOfIterations() int {
	return numberOfIterations
}
