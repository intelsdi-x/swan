package provisioner

import (
	"os/exec"
	"sync"
	"fmt"
)

type Shell struct {
	commands []*Command
}

func NewShell(commands []*Command) *Shell {
	return &Shell{
		commands: commands,
	}
}

func (s *Shell) LenCommands() int {
	return len(s.commands)
}

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
