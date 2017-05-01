package experiment

// Range allows to iterate over values it specifies
type Range struct {
	From float64
	To   float64
	Step float64
}

// Iterate implements Iterator interface
func (r Range) Iterate(runnable interface{}) {
	for i := r.From; i < r.To; i = i + r.Step {
		Call(runnable, i)
	}
}
