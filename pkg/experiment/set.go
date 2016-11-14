package experiment

// Set wllows to iterate over slice of values of any type
type Set []interface{}

// Iterate implements Iterator interface
func (s Set) Iterate(runnable interface{}) {
	for _, v := range s {
		Call(runnable, v)
	}
}
