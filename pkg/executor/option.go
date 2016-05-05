package executor

// Option is monad for Optional variable in Go. If Option is not filled then it is None.
// Inspired with: github.com/SimonRichardson/wishful/blob/master/useful/option.go
type Option interface {
	// Empty checks if Option has value.
	Empty() bool
	// Get gets the value. It returns panic when empty.
	Get() interface{}
	// GetOrElse gets the value. It returns else value when empty.
	GetOrElse(interface{}) interface{}
}

// NoneOption does not contain value and it's empty.
type NoneOption struct{}

// None constructor.
func None() NoneOption {
	return NoneOption{}
}

// Empty checks if Option has value.
func (n NoneOption) Empty() bool {
	return true
}

// GetOrElse gets the value. It returns else value when empty.
func (n NoneOption) GetOrElse(x interface{}) interface{} {
	return x
}

// Get gets the value. It returns panic when empty.
func (n NoneOption) Get() interface{} {
	panic("Option is empty.")
}

// SomeOption contains value.
type SomeOption struct {
	x interface{}
}

// Some constructor.
func Some(x interface{}) SomeOption {
	return SomeOption{x: x}
}

// Empty checks if Option has value.
func (s SomeOption) Empty() bool {
	return false
}

// Get gets the value. It returns panic when empty.
func (s SomeOption) Get() interface{} {
	return s.x
}

// GetOrElse gets the value. It returns else value when empty.
func (s SomeOption) GetOrElse(x interface{}) interface{} {
	return s.x
}
