package topo

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

// Thread represents a hyperthread, typically presented by the operating
// system as a logical CPU.
type Thread interface {
	ID() int
	Core() int
	Socket() int
	Equals(Thread) bool
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
		logrus.Errorf("NewThreadFromID: Could not discover CPUs\n")
		return thread, err
	}

	foundThreads := allThreads.Filter(
		func(t Thread) bool {
			return t.ID() == id
		})

	if len(foundThreads) == 0 {
		logrus.Errorf("NewThreadFromID: Could not find thread with id %d on platform\n", id)
		return thread, fmt.Errorf("could not find thread with id %d on platform", id)
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

func (t thread) Equals(that Thread) bool {
	return t.ID() == that.ID() &&
		t.Core() == that.Core() &&
		t.Socket() == that.Socket()
}
