package experiment

import (
	"strconv"
	"strings"
)

type Experimentable func(value interface{})

// Range allows to run a function across a range of integers
func Range(rangeSpec string, r Experimentable) {
	from, to := parseRangeSpec(rangeSpec)
	for i := from; i <= to; i++ {
		r(i)
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
		r(v)
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
func Permute(specs []string, r Experimentable) {
	if len(specs) > 1 {
		recursive := func(value interface{}) {
			Permute(specs[1:], r)
		}
		if isSetSpec(specs[0]) {
			Set(specs[0], recursive)
		} else if isRangeSpec(specs[0]) {
			Range(specs[0], recursive)
		}
	} else {
		if isSetSpec(specs[0]) {
			Set(specs[0], r)
		} else if isRangeSpec(specs[0]) {
			Range(specs[0], r)
		}

	}
}

func isSetSpec(spec string) bool {
	return strings.Contains(spec, ",")
}

func isRangeSpec(spec string) bool {
	return strings.Contains(spec, "-")
}
