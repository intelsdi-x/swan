package dummy

// Dummy class serves as fill code to exercise test and linting.
type Dummy struct {}

// NewDummy returns a Dummy object.
func NewDummy() *Dummy {
	return &Dummy{}
}

// Foo does nothing else by returning 42.
func (d *Dummy) Foo() int {
	return 42
}
