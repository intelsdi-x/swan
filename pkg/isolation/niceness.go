package isolation

import "fmt"

// Nice is a Decorator that allows to set process priority before starting it.
type Nice struct {
	Niceness int
}

// Decorate implements Decorator interface.
func (n Nice) Decorate(command string) string {
	return fmt.Sprintf("nice --adjustment %d %s", n.Niceness, command)
}
