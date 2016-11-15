package experiment

import (
	"fmt"
	"reflect"
)

// Iterator is an interface that struct needs to implement in order to be part of Arg to allow Iterate() to consume it.
type Iterator interface {
	Iterate(interface{})
}

// experimentContext stores current items from all iterations. Call() handles its value.
var experimentContext []interface{}

// numberOfIterations stores number of calls to DryRun().
var numberOfIterations int

// Arg represents named iteration that Permutate() and Iterate() can consume.
type Arg struct {
	Name string
	Iterator
}

// Permute accepts variable number of arguments. Allows to specify multiple instances of Arg of various type of Spec. The last one needs to be a function that is to be executed.
func Permute(args ...interface{}) {
	if len(args) > 2 {
		// More then two arguments indicates that we need to call Permutate() recursively. In order to do so we create a closure that will be passed to Iterate().
		recursive := func() {
			Permute(args[1:]...)
		}
		Iterate(args[0].(Arg), recursive)
	} else if len(args) == 2 {
		// If there are two argument only then we are on the leaf and have nowhere to recurse to.
		Iterate(args[0].(Arg), args[1])
	}
}

// Iterate is a public interface of the experiment. It can recognize type of iteration and runs approptiate code. Last argument needs to be a function that is to be executed.
func Iterate(arg Arg, runnable interface{}) {
	fmt.Printf("Executing step %s with arguments %s\n", arg.Name, arg.Iterator)
	arg.Iterate(runnable)
}

// Call handles calling a function passed to Iterate(). It is capable of calling a function with no arguments or with number of arguments equal to number of Arg instances parsed.
func Call(runnable, localContext interface{}) {
	experimentContext = append(experimentContext, localContext)
	defer func() { experimentContext = experimentContext[:len(experimentContext)-1] }()
	function := reflect.ValueOf(runnable)
	argsIn := function.Type().NumIn()
	if argsIn == len(experimentContext) {
		// number of runnable arguments should be equal to current context size (including localContext)...
		var args []reflect.Value
		for _, v := range experimentContext {
			args = append(args, reflect.ValueOf(v))
		}
		function.Call(args)
	} else if argsIn == 0 {
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
