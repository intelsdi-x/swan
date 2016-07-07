package isolation

import "fmt"

type nice struct {
	value int
}

// NewNice creates new of "nice prepending decorator" for value != 0.
func NewNice(value int) Decorator {
	return &nice{value: value}
}

// Decorate prepare "nice -n" prefixed command for value other than 0.
func (n *nice) Decorate(command string) (decorated string) {
	if n.value != 0 {
		return fmt.Sprintf("nice -n %d %s", n.value, command)
	}
	return command
}
