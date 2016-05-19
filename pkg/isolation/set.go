package isolation

import (
	"reflect"
	"strconv"
	"strings"
)

// Set represents a traditional set type so we can do intersections, joins, etc.
// on core and memory node ids.
type Set map[int]struct{}

// Empty returns true iff this set has exactly zero elements.
func (as Set) Empty() bool {
	return len(as) == 0
}

// Contains returns true if the supplied item is present in this set.
func (as Set) Contains(item int) bool {
	_, found := as[item]
	return found
}

// Add mutates this set to include the supplied item.
func (as Set) Add(item int) {
	as[item] = struct{}{}
}

// Remove mutates this set to remove the supplied item.
func (as Set) Remove(item int) {
	delete(as, item) // does nothing if item does not exist
}

// Equals returns true iff the supplied set is equal to this set.
func (as Set) Equals(bs Set) bool {
	return reflect.DeepEqual(as, bs)
}

// Union returns a new set that contains all of the items from this set
// and all of the items from the supplied set.
// It does not mutate either set.
func (as Set) Union(bs Set) Set {
	result := NewSet()
	for a := range as {
		result.Add(a)
	}
	for b := range bs {
		result.Add(b)
	}
	return result
}

// Intersection returns a new set that contains all of the items that are
// present in both this set and the supplied set.
// It does not mutate either set.
func (as Set) Intersection(bs Set) Set {
	result := NewSet()
	for a := range as {
		if bs.Contains(a) {
			result.Add(a)
		}
	}
	return result
}

// Difference returns a new set that contains all of the items that are
// present in this set and not the supplied set.
// It does not mutate either set.
func (as Set) Difference(bs Set) Set {
	result := NewSet()
	for a := range as {
		result.Add(a)
	}
	for b := range bs {
		result.Remove(b)
	}
	return result
}

// NewSet returns a set containing the element from ids.
func NewSet(ids ...int) Set {
	ret := Set{}
	for _, id := range ids {
		ret.Add(id)
	}
	return ret
}

// NewSetFromRange creates a set from traditional cgroup set representation.
// For example, "0-5,34,46-48"
func NewSetFromRange(rangesString string) (Set, error) {
	ret := Set{}

	// Split ranges string:
	// "0-5,34,46-48" becomes ["0-5", "34", "46-48"]
	ranges := strings.Split(rangesString, ",")
	for _, r := range ranges {
		boundaries := strings.Split(r, "-")

		if len(boundaries) == 1 {
			// Some entries may only contain one id, like "34" in our example.
			id, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return Set{}, err
			}

			ret[id] = struct{}{}
		} else if len(boundaries) == 2 {
			// For ranges, we parse start and end.
			start, err := strconv.Atoi(boundaries[0])
			if err != nil {
				return Set{}, err
			}

			end, err := strconv.Atoi(boundaries[1])
			if err != nil {
				return Set{}, err
			}

			// And add all the ids to the set.
			// Here, "0-5", "46-48" should become [0,1,2,3,4,5,46,47,48]
			for id := start; id <= end; id++ {
				ret.Add(id)
			}
		}
	}

	return ret, nil
}
