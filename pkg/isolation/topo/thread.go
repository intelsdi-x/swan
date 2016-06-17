package topo

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
