package provisioner

// Command encapsulated a command to be executed.
// Instead of using a flat string, this struct is meant to contain
// details which user to run as, permission, work directory, etc.
type Command struct {
	cmd string
}

// NewCommand returns a command instance.
func NewCommand(cmd string) Command {
	command := Command {
		cmd: cmd,
	}
	return command
}
