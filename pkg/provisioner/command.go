package provisioner

// Command encapsulated a command to be executed.
// Instead of using a flat string, this struct is meant to contain
// details which user to run as, permission, work directory, etc.
type Command struct {
	command string
}

// NewCommand returns a command instance.
func NewCommand(command string) *Command {
	return &Command{
		command: command,
	}
}

// CommandString returns the command to run.
func (c *Command) CommandString() string {
	return c.command
}
