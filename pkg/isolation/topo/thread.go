package topo

import (
	"fmt"
)

// Thread represents a hyperthread, typically presented by the operating
// system as a logical CPU.
type Thread interface {
	ID() int
	Core() int
	Socket() int
}

// NewThread returns a new thread with the supplied thread, core, and
// socket IDs.
func NewThread(id int, core int, socket int) Thread {
	return thread{id, core, socket}
}

// NewThreadFromID returns new Thread from ThreadID
func NewThreadFromID(id int) (thread Thread, err error) {
	allThreads, err := Discover()
	if err != nil {
		return thread, err
	}

	foundThreads := allThreads.Filter(func(t Thread) bool {
		return t.ID() == id
	})

	if len(foundThreads) == 0 {
		return thread, fmt.Errorf("Could not find thread with id %d on platform", id)
	}

	return foundThreads[0], err
}

type thread struct {
	id     int
	core   int
	socket int
}

func (t thread) ID() int {
	return t.id
}

func (t thread) Core() int {
	return t.core
}

func (t thread) Socket() int {
	return t.socket
}
