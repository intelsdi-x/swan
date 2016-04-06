package provisioning

import "syscall"

// Status represents the exit status for a command.
type Status struct {
	code syscall.WaitStatus
 	output string
}
