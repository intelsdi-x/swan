package experiment

// List allows to iterate over slice of values of any type
type List []interface{}

// Iterate implements Iterator interface
func (s List) Iterate(runnable interface{}) {
	for _, v := range s {
		Call(runnable, v)
	}
}
