package experiment

import (
	"reflect"
	"strconv"
	"strings"
)

var experimentContext []interface{}

// Range allows to run a function across a range of integers
func Range(specs ...interface{}) {
	from, to := parseRangeSpec(specs[0])
	for i := from; i <= to; i++ {
		call(specs[1], i)
	}
}

func parseRangeSpec(rangeSpec interface{}) (from, to int) {
	boundaries := strings.Split(rangeSpec.(string), "-")
	from, _ = strconv.Atoi(boundaries[0])
	to, _ = strconv.Atoi(boundaries[1])

	return from, to
}

func call(r, localContext interface{}) {
	experimentContext = append(experimentContext, localContext)
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
	experimentContext = experimentContext[:len(experimentContext)-1]
}

// Set allows to run a function across set of values
func Set(specs ...interface{}) {
	set := parseSetSpec(specs[0])
	for _, v := range set {
		call(specs[1], v)
	}
}

func parseSetSpec(setSpec interface{}) (set []interface{}) {
	items := strings.Split(setSpec.(string), ",")
	for _, v := range items {
		set = append(set, v)
	}

	return set
}

// Permutate accepts slices of specs, tell set from range and run them recursively
func Permutate(specs ...interface{}) {
	if len(specs) > 2 {
		recursive := func() {
			Permutate(specs[1:]...)
		}
		if isSetSpec(specs[0]) {
			Set(specs[0].(string), recursive)
		} else if isRangeSpec(specs[0]) {
			Range(specs[0].(string), recursive)
		}
	} else {
		if isSetSpec(specs[0]) {
			Set(specs[0].(string), specs[1])
		} else if isRangeSpec(specs[0]) {
			Range(specs[0].(string), specs[1])
		}

	}
}

func isSetSpec(spec interface{}) bool {
	return strings.Contains(spec.(string), ",")
}

func isRangeSpec(spec interface{}) bool {
	return strings.Contains(spec.(string), "-")
}
