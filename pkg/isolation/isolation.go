package isolation

type Isolation interface{
	Init() error
	Perform(taskPid TaskPID) error
}
