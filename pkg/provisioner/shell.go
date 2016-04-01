package provisioner

import (
	"fmt"
	"github.com/hypersleep/easyssh"
	"os/exec"
	"sync"
)

type Shell struct {
	commands []Command
}

// NewShell returns a Shell instance from a set of commands.
func NewShell(commands []Command) *Shell {
	return &Shell{
		commands: commands,
	}
}

// LenCommands returns the number of commands to be executed.
func (s *Shell) CommandsCount() int {
	return len(s.commands)
}

// Execute runs the commands given during Shell construction in
// parallel. The returned channel will only complete when all commands
// have completed.
func (s *Shell) Execute() <-chan []Status {
	out := make(chan []Status)
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(s.commands))
		for _, command := range(s.commands) {

			go func(){
				defer wg.Done()

				cmd := exec.Command("sh", "-c", command.cmd)
				cmd.Run()
				fmt.Printf("Ended command\n")
			}()
		}
		wg.Wait()
		out <- []Status{}
	}()
	return out
}

func (s *Shell) ExecuteRemotely() <-chan []Status {
        out := make(chan []Status)
	ssh := &easyssh.MakeConfig{
		User:   "root",
		Server: "localhost",
		Key:    "/.ssh/id_rsa",
		Port:   "22",
	}
	go func() {
		var wg sync.WaitGroup
                statuses := make([]Status, s.CommandsCount())
		wg.Add(len(s.commands))
		for _, command := range s.commands {

			go func() {
				defer wg.Done()
				response, err := ssh.Run(command.cmd)
				if err != nil {
					panic("Can't run remote command: " + err.Error())
				} else {
				        statuses = append(statuses, Status{response})
				}
				defer wg.Done()
			}()
		}
		out <- statuses
		wg.Wait()
	}()
	return out
}
