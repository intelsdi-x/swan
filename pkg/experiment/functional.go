package experiment

import (
	"strconv"
	"strings"
)

var experimentContext []interface{}

// Experimentable is a function that Range(), Set() and Permute() can consume
type Experimentable func(...interface{})

// Range allows to run a function across a range of integers
func Range(rangeSpec string, r Experimentable) {
	from, to := parseRangeSpec(rangeSpec)
	for i := from; i <= to; i++ {
		experimentContext := append(experimentContext, i)
		r(experimentContext...)
		experimentContext = experimentContext[:len(experimentContext)-1]
	}
}

func parseRangeSpec(rangeSpec string) (from, to int) {
	boundaries := strings.Split(rangeSpec, "-")
	from, _ = strconv.Atoi(boundaries[0])
	to, _ = strconv.Atoi(boundaries[1])

	return from, to
}

// Set allows to run a function across set of values
func Set(setSpec string, r Experimentable) {
	set := parseSetSpec(setSpec)
	for _, v := range set {
		experimentContext = append(experimentContext, v)
		r(experimentContext...)
		experimentContext = experimentContext[:len(experimentContext)-1]
	}
}

func parseSetSpec(setSpec string) (set []interface{}) {
	items := strings.Split(setSpec, ",")
	for _, v := range items {
		set = append(set, v)
	}

	return set
}

// Permute accepts slices of specs, tell set from range and run them recursively
func Permutate(specs ...interface{}) {
	if len(specs) > 2 {
		recursive := func(context ...interface{}) {
			experimentContext := append(experimentContext, context...)
			Permutate(specs[1:]...)
			experimentContext = experimentContext[:len(experimentContext)-1]
		}
		if isSetSpec(specs[0]) {
			Set(specs[0].(string), recursive)
		} else if isRangeSpec(specs[0]) {
			Range(specs[0].(string), recursive)
		}
	} else {
		if isSetSpec(specs[0]) {
			Set(specs[0].(string), specs[1].(func(...interface{})))
		} else if isRangeSpec(specs[0]) {
			Range(specs[0].(string), specs[1].(func(...interface{})))
		}

	}
}

func isSetSpec(spec interface{}) bool {
	return strings.Contains(spec.(string), ",")
}

func isRangeSpec(spec interface{}) bool {
	return strings.Contains(spec.(string), "-")
}
