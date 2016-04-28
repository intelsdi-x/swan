package isolation

//Isolation of resources exposes these interfaces
type Isolation interface{
 Isolate(PID int) error
 Delete() error
}

