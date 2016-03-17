package shell

import (
	"os/exec"
	"sync"
	"fmt"
)

// Shell represents a set of commands to be run in parallel.
type Shell struct {
	commands []*Command
}

// NewShell returns a Shell instance from a set of commands.
func NewShell(commands []*Command) *Shell {
	return &Shell{
		commands: commands,
	}
}

// LenCommands returns the number of commands to be executed.
func (s *Shell) LenCommands() int {
	return len(s.commands)
}

// Execute runs the commands given during Shell construction in
// parallel. The returned channel will only complete when all commands
// have completed.
func (s *Shell) Execute() <-chan []*Status {
	out := make(chan []*Status)
	go func() {
		var wg sync.WaitGroup

		for _, command := range(s.commands) {
			wg.Add(1)
			go func(){
				defer wg.Done()

				cmd := exec.Command("sh", "-c", command.CommandString())
				cmd.Start()

				fmt.Printf("Started command: '%s'\n", command.CommandString())
				cmd.Wait()
				fmt.Printf("Ended command\n")
			}()
		}
		wg.Wait()
		out <- []*Status{}
	}()
	return out
}
