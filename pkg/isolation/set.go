package isolation

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// IntSet represents a traditional set type so we can do intersections, joins, etc.
// on core and memory node ids.
type IntSet map[int]struct{}

// Empty returns true iff this set has exactly zero elements.
func (s IntSet) Empty() bool {
	return len(s) == 0
}

// Contains returns true if the supplied element is present in this set.
func (s IntSet) Contains(elem int) bool {
	_, found := s[elem]
	return found
}

// Add mutates this set to include the supplied element.
func (s IntSet) Add(elem int) {
	s[elem] = struct{}{}
}

// Remove mutates this set to remove the supplied element.
func (s IntSet) Remove(elem int) {
	delete(s, elem) // does nothing if item does not exist
}

// Equals returns true iff the supplied set is equal to this set.
func (s IntSet) Equals(t IntSet) bool {
	return reflect.DeepEqual(s, t)
}

// Subset returns true iff the supplied set contains all the elements in
// this set (e.g. s is a subset of t).
func (s IntSet) Subset(t IntSet) bool {
	result := true
	for elem := range s {
		if !t.Contains(elem) {
			result = false
			break
		}
	}
	return result
}

// Union returns a new set that contains all of the elements from this set
// and all of the elements from the supplied set.
// It does not mutate either set.
func (s IntSet) Union(t IntSet) IntSet {
	result := NewIntSet()
	for elem := range s {
		result.Add(elem)
	}
	for elem := range t {
		result.Add(elem)
	}
	return result
}

// Intersection returns a new set that contains all of the elements that are
// present in both this set and the supplied set.
// It does not mutate either set.
func (s IntSet) Intersection(t IntSet) IntSet {
	result := NewIntSet()
	for elem := range s {
		if t.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

// Difference returns a new set that contains all of the elements that are
// present in this set and not the supplied set.
// It does not mutate either set.
func (s IntSet) Difference(t IntSet) IntSet {
	result := NewIntSet()
	for elem := range s {
		result.Add(elem)
	}
	for elem := range t {
		result.Remove(elem)
	}
	return result
}

// Take returns a new set containing n items from this set. If n is greater
// than the number of elements in the set, returns an error.
func (s IntSet) Take(n int) (IntSet, error) {
	if n > len(s) {
		return nil, errors.Errorf("cannot take %d elements from a set of size %d", n, len(s))
	}
	result := NewIntSet()
	for _, elem := range s.AsSlice()[0:n] {
		result.Add(elem)
	}
	return result, nil
}

// AsSlice returns a slice of integers that contains all elements from
// this set.
func (s IntSet) AsSlice() []int {
	result := []int{}
	for elem := range s {
		result = append(result, elem)
	}
	return result
}

// AsRangeString returns a traditional cgroup set representation.
// For example, "0,1,2,3,4,5,34,46,47,48"
func (s IntSet) AsRangeString() string {
	elemStrs := []string{}
	for elem := range s {
		elemStrs = append(elemStrs, strconv.Itoa(elem))
	}
	return strings.Join(elemStrs, ",")
}

// NewIntSet returns a new set containing all of the supplied elements.
func NewIntSet(elems ...int) IntSet {
	result := IntSet{}
	for _, elem := range elems {
		result.Add(elem)
	}
	return result
}

// NewIntSetFromRange creates a set from traditional cgroup set representation.
// For example, "0-5,34,46-48"
func NewIntSetFromRange(rangesString string) (IntSet, error) {
	result := NewIntSet()

	if rangesString == "" {
		return result, nil
	}

	// Split ranges string:
	// "0-5,34,46-48" becomes ["0-5", "34", "46-48"]
	ranges := strings.Split(rangesString, ",")
	for _, r := range ranges {
		boundaries := strings.Split(r, "-")

		if len(boundaries) == 1 {
			// Some entries may only contain one id, like "34" in our example.
			elem, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return nil, errors.Wrapf(err, "could not atoi %q", boundaries[0])
			}
			result.Add(elem)
		} else if len(boundaries) == 2 {
			// For ranges, we parse start and end.
			start, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return nil, errors.Wrapf(err, "could not atoi %q", boundaries[0])
			}

			end, err := strconv.Atoi(boundaries[1])
			if err != nil {
				return nil, errors.Wrapf(err, "could not atoi %q", boundaries[1])
			}

			// And add all the ids to the set.
			// Here, "0-5", "46-48" should become [0,1,2,3,4,5,46,47,48]
			for elem := start; elem <= end; elem++ {
				result.Add(elem)
			}
		}
	}

	return result, nil
}
