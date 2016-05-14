package isolation

//Isolation of resources exposes these interfaces
type Isolation interface {
	Create() error
	Isolate(PID int) error
	Clean() error
	Path() string
	Controller() string
}
