package topo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/intelsdi-x/swan/pkg/isolation"
)

// ThreadSet represents a subset of the available hyperthreads on a system.
type ThreadSet []Thread

// NewThreadSet returns a newly allocated thread set.
func NewThreadSet() ThreadSet {
	return []Thread{}
}

// NewThreadSetFromIntSet returns newly allocated thread set from IntSet with Thread IDs.
func NewThreadSetFromIntSet(threads isolation.IntSet) (threadSet ThreadSet, err error) {
	for threadID := range threads {
		thread, err := NewThreadFromID(threadID)
		if err != nil {
			fmt.Errorf("Could not create thread with ID %d: %s", threadID, err)
		}
		threadSet = append(threadSet, thread)
	}
	return threadSet, err
}

// Partition returns two newly allocated thread sets: the first contains
// threads from this set that match the supplied predicate and the second
// contains threads that do not.
func (s ThreadSet) Partition(by func(Thread) bool) (ThreadSet, ThreadSet) {
	left := ThreadSet{}
	right := ThreadSet{}
	for _, t := range s {
		if by(t) {
			left = append(left, t)
		} else {
			right = append(right, t)
		}
	}
	return left, right
}

// Filter returns a newly allocated thread set containing all elements
// from this set that match the supplied predicate.
func (s ThreadSet) Filter(by func(Thread) bool) ThreadSet {
	res := ThreadSet{}
	for _, t := range s {
		if by(t) {
			res = append(res, t)
		}
	}
	return res
}

// AvailableThreads returns the set of thread ids for threads in this
// thread set.
func (s ThreadSet) AvailableThreads() isolation.IntSet {
	threads := isolation.NewIntSet()
	for _, t := range s {
		threads.Add(t.ID())
	}
	return threads
}

// AvailableCores returns the set of core ids for threads in this
// thread set.
func (s ThreadSet) AvailableCores() isolation.IntSet {
	cores := isolation.NewIntSet()
	for _, t := range s {
		cores.Add(t.Core())
	}
	return cores
}

// AvailableSockets returns the set of socket ids for threads in this
// thread set.
func (s ThreadSet) AvailableSockets() isolation.IntSet {
	sockets := isolation.NewIntSet()
	for _, t := range s {
		sockets.Add(t.Socket())
	}
	return sockets
}

// Threads returns a newly allocated thread set containing `n` distinct
// threads from this thread set. If there are fewer than `n` available,
// returns an error.
func (s ThreadSet) Threads(n int) (ThreadSet, error) {
	threads, err := s.AvailableThreads().Take(n)
	if err != nil {
		return nil, err
	}
	return s.Filter(func(t Thread) bool { return threads.Contains(t.ID()) }), nil
}

// Cores returns a newly allocated thread set containing all threads from
// `n` distinct cores. If there are fewer than `n` available, returns an error.
func (s ThreadSet) Cores(n int) (ThreadSet, error) {
	cores, err := s.AvailableCores().Take(n)
	if err != nil {
		return nil, err
	}
	return s.Filter(func(t Thread) bool { return cores.Contains(t.Core()) }), nil
}

// Sockets returns a newly allocated thread set containing all threads from
// `n` distinct sockets. If there are fewer than `n` available, returns
// an error.
func (s ThreadSet) Sockets(n int) (ThreadSet, error) {
	sockets, err := s.AvailableSockets().Take(n)
	if err != nil {
		return nil, err
	}
	return s.Filter(func(t Thread) bool { return sockets.Contains(t.Socket()) }), nil
}

// FromThreads returns a newly allocated thread set containing all threads with
// the supplied IDs. If any of the supplied thread IDs are invalid, returns
// an error.
func (s ThreadSet) FromThreads(threadIDs ...int) (ThreadSet, error) {
	threads := isolation.NewIntSet(threadIDs...)
	if !threads.Subset(s.AvailableThreads()) {
		return nil, fmt.Errorf("invalid thread id(s): available threads are %s", s.AvailableThreads().AsRangeString())
	}
	return s.Filter(func(t Thread) bool { return threads.Contains(t.ID()) }), nil
}

// FromCores returns a newly allocated thread set containing all threads from
// the supplied cores. If any of the supplied cores are invalid, returns
// an error.
func (s ThreadSet) FromCores(coreIDs ...int) (ThreadSet, error) {
	cores := isolation.NewIntSet(coreIDs...)
	if !cores.Subset(s.AvailableCores()) {
		return nil, fmt.Errorf("invalid core id(s): available cores are %s", s.AvailableCores().AsRangeString())
	}
	return s.Filter(func(t Thread) bool { return cores.Contains(t.Core()) }), nil
}

// FromSockets returns a newly allocated thread set containing all threads from
// the supplied sockets. If any of the supplied sockets are invalid, returns
// an error.
func (s ThreadSet) FromSockets(socketIDs ...int) (ThreadSet, error) {
	sockets := isolation.NewIntSet(socketIDs...)
	if !sockets.Subset(s.AvailableSockets()) {
		return nil, fmt.Errorf("invalid socket id(s): available sockets are %s", s.AvailableSockets().AsRangeString())
	}
	return s.Filter(func(t Thread) bool { return sockets.Contains(t.Socket()) }), nil
}

// Contains returns true iff this set contains the supplied thread.
func (s ThreadSet) Contains(t Thread) bool {
	for _, elem := range s {
		if elem.Equals(t) {
			return true
		}
	}
	return false
}

// Difference returns a newly allocated thread set containing all threads from
// this set except those contained in the supplied set.
func (s ThreadSet) Difference(ts ThreadSet) ThreadSet {
	return s.Filter(func(t Thread) bool { return !ts.Contains(t) })
}

// Remove returns a newly allocated thread set containing all threads from
// this set except the supplied thread.
func (s ThreadSet) Remove(t Thread) ThreadSet {
	return s.Filter(func(elem Thread) bool { return !elem.Equals(t) })
}

// RemoveThreadSet Remove returns a newly allocated thread set containing all threads from
// this set except the supplied threads.
func (s ThreadSet) RemoveThreadSet(ts ThreadSet) (result ThreadSet) {
	result = s
	for _, thread := range ts {
		result = result.Remove(thread)
	}
	return result
}

// ToCPUSetNotation returns thread IDs in cpuset notation.
// Example return: 0,1,2,3,4 (...)
func (s ThreadSet) ToCPUSetNotation() string {
	threads := make([]string, len(s))
	for idx, thread := range s {
		threads[idx] = strconv.Itoa(thread.ID())
	}
	return strings.Join(threads, ",")
}
