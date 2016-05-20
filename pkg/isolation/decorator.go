package isolation

//Decorator allows to decorate launcher output.
type Decorator interface {
	Decorate(string) string
}

//Decorators represents array of Decorator implementations.
type Decorators []Decorator

//Decorate uses all available decorators to modify the command
//(and implements Decorator interface).
func (d Decorators) Decorate(command string) string {
	for _, decorator := range d {
		command = decorator.Decorate(command)
	}
	return command
}
