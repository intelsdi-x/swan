package experiment

// TBD

// Parameters of isolation that consist of:
// Isolation mechanism, which has to be unequivocal, precise and closed;
// (resources (like cpu, memory) assigned to the job (e.g CPU pinning,
// pinning NUMA nodes, CAT / MBA) - mostly using cgroups)
// defines topology
type Isolation struct {
}

func (i Isolation) String() string {
	return "Isolation object not defined"
}
