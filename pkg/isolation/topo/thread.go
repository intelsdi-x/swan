package topo

import "github.com/Sirupsen/logrus"

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

func NewThreadFromId(id int) Thread {
	allThreads, err := Discover()
	if err != nil {
		logrus.Fatal(err)
	}

	foundThread := allThreads.Filter(func(t Thread) bool {
		return t.ID() == id
	})

	return foundThread[0]
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
