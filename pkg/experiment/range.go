package experiment

// Interval allows to iterate over values it specifies
type Interval struct {
	From float64
	To   float64
	Step float64
}

// Execute iterates over values that Interval
func (r *Interval) Execute(runnable interface{}) {
	for i := r.From; i < r.To; i = i + r.Step {
		call(runnable, i)
	}
}
