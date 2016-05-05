package executor

// OptionInt is monad for Optional integer variable in Go. If Option is not filled then it is None.
// Inspired with: github.com/SimonRichardson/wishful/blob/master/useful/option.go
type OptionInt interface {
	// Empty checks if Option has value.
	Empty() bool
	// Get gets the value. It returns panic when empty.
	Get() int
	// GetOrElse gets the value. It returns else value when empty.
	GetOrElse(int) int
}

// NoneIntOption does not contain value and it's empty.
type NoneIntOption struct{}

// NoneInt option constructor.
func NoneInt() NoneIntOption {
	return NoneIntOption{}
}

// Empty checks if Option has value.
func (n NoneIntOption) Empty() bool {
	return true
}

// GetOrElse gets the value. It returns else value when empty.
func (n NoneIntOption) GetOrElse(x int) int {
	return x
}

// Get gets the value. It returns panic when empty.
func (n NoneIntOption) Get() int {
	panic("Option is empty.")
}

// SomeIntOption contains value.
type SomeIntOption struct {
	x int
}

// SomeInt option constructor.
func SomeInt(x int) SomeIntOption {
	return SomeIntOption{x: x}
}

// Empty checks if Option has value.
func (s SomeIntOption) Empty() bool {
	return false
}

// Get gets the value. It returns panic when empty.
func (s SomeIntOption) Get() int {
	return s.x
}

// GetOrElse gets the value. It returns else value when empty.
func (s SomeIntOption) GetOrElse(x int) int {
	return s.x
}
