package isolation

type Isolation interface{
	Init(targetHost string) error
	Perform(taskPid int)
}
