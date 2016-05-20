package isolation

//Isolation of resources exposes these interfaces
type Isolation interface {
	Decorator
	Create() error
	Isolate(PID int) error
	Clean() error
}
