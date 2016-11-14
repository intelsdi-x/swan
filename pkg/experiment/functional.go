package experiment

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var experimentContext []interface{}

var numberOfIterations int

type Arg struct {
	Name string
	Spec interface{}
}

// Permutate accepts slices of specs, tell set from range and run them recursively
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

// Iterate is a public interface of the experiment. It can recognize type of iteration and runs approptiate code.
func Iterate(specs ...interface{}) {
	fmt.Printf("Executing step %s with arguments %s\n", specs[0].(Arg).Name, specs[0].(Arg).Spec.(string))
	if isSetSpec(specs[0]) {
		set(specs...)
	} else if isIntervalSpec(specs[0]) {
		interval(specs...)
	}
}

func isSetSpec(spec interface{}) bool {
	return strings.Contains(spec.(Arg).Spec.(string), ",")
}

func set(specs ...interface{}) {
	set := parseSetSpec(specs[0])
	for _, v := range set {
		call(specs[1], v)
	}
}

func parseSetSpec(setSpec interface{}) (set []interface{}) {
	items := strings.Split(setSpec.(Arg).Spec.(string), ",")
	for _, v := range items {
		set = append(set, v)
	}

	return set
}

func call(r, localContext interface{}) {
	experimentContext = append(experimentContext, localContext)
	defer func() { experimentContext = experimentContext[:len(experimentContext)-1] }()
	function := reflect.ValueOf(r)
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

func isIntervalSpec(spec interface{}) bool {
	return strings.Contains(spec.(Arg).Spec.(string), "-")
}

func interval(specs ...interface{}) {
	from, to := parseIntervalSpec(specs[0])
	for i := from; i <= to; i++ {
		call(specs[1], i)
	}
}

func parseIntervalSpec(rangeSpec interface{}) (from, to int) {
	boundaries := strings.Split(rangeSpec.(Arg).Spec.(string), "-")
	from, _ = strconv.Atoi(boundaries[0])
	to, _ = strconv.Atoi(boundaries[1])

	return from, to
}

// DryRun is a helper function that allows to estimate number of iterations that Permutate() generates
func DryRun() {
	numberOfIterations++
}

// GetNumberOfIterations retrieves number of dry run iterations
func GetNumberOfIterations() int {
	return numberOfIterations
}
