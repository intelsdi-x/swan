package experiment

import (
	"fmt"
	"reflect"
)

var experimentContext []interface{}

var numberOfIterations int

type Arg struct {
	Name string
	Spec interface{}
}

// Permutate accepts variable number of arguments. Allows to specify multiple instances of Arg of various type of Spec. The last one needs to be a function that is to be executed.
func Permutate(specs ...interface{}) {
	if len(specs) > 2 {
		recursive := func() {
			Permutate(specs[1:]...)
		}
		Iterate(specs[0], recursive)
	} else {

		Iterate(specs...)
	}
}

// Iterate is a public interface of the experiment. It can recognize type of iteration and runs approptiate code. Last argument needs to be a function that is to be executed.
func Iterate(specs ...interface{}) {
	fmt.Printf("Executing step %s with arguments %s\n", specs[0].(Arg).Name, specs[0].(Arg).Spec)
	if isSetSpec(specs) {
		set(specs[0].(Arg), specs[1])
	} else if isIntervalSpec(specs) {
		interval(specs[0].(Arg), specs[1])
	}
}

func isSetSpec(specs []interface{}) bool {
	value := reflect.ValueOf(specs[0].(Arg).Spec)

	return value.Kind() == reflect.Slice && len(specs) == 2
}

func set(spec Arg, runnable interface{}) {
	for _, v := range spec.Spec.([]interface{}) {
		call(runnable, v)
	}
}

// call handles calling a function passed to Iterate(). It is capable of calling a function with no arguments or with number of arguments equal to number of Arg instances parsed.
func call(runnable, localContext interface{}) {
	experimentContext = append(experimentContext, localContext)
	defer func() { experimentContext = experimentContext[:len(experimentContext)-1] }()
	function := reflect.ValueOf(runnable)
	if function.Type().NumIn() > 0 {
		var args []reflect.Value
		for _, v := range experimentContext {
			args = append(args, reflect.ValueOf(v))
		}
		function.Call(args)
	} else {

		function.Call(nil)
	}
}

func isIntervalSpec(specs []interface{}) bool {
	_, ok := specs[0].(Arg).Spec.(*Interval)

	return ok && len(specs) == 2
}

func interval(spec Arg, runnable interface{}) {
	spec.Spec.(*Interval).Execute(runnable)
}

// DryRun is a helper function that allows to estimate number of iterations that Permutate() generates
func DryRun() {
	numberOfIterations++
}

// GetNumberOfIterations retrieves number of dry run iterations
func GetNumberOfIterations() int {
	return numberOfIterations
}
