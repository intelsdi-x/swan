package provisioning

type Task struct{
	id int
	command string
}

// NewTask returns a Task instance.
func NewTask(id int, command string) Task {
	t := Task{
		id,
		command,
	}
	return t
}