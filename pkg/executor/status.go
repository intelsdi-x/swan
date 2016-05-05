package executor

// Status represents the status of a task in the current point of time.
// TODO: As Niklas mentioned we need to remove that in next PR and stay with exitCode only.
type Status struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
