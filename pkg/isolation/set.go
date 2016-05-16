package isolation

import (
	"strconv"
	"strings"
)

// Set represents a traditional set type so we can do intersections, joins, etc.
// on core and memory node ids.
type Set map[int]struct{}

// NewSet returns a set containing the element from ids.
func NewSet(ids ...int) Set {
	ret := Set{}
	for _, id := range ids {
		ret[id] = struct{}{}
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
				ret[id] = struct{}{}
			}
		}
	}

	return ret, nil
}
