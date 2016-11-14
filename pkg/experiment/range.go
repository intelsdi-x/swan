package experiment

// Interval allows to iterate over values it specifies
type Interval struct {
	From float64
	To   float64
	Step float64
}

// Iterate implements Iterator interface
func (r *Interval) Iterate(runnable interface{}) {
	for i := r.From; i < r.To; i = i + r.Step {
		Call(runnable, i)
	}
}
