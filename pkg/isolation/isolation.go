package isolation


type Isolation interface{
 Isolate(PID int) error
 Clean() error
}

