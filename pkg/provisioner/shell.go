package provisioner

import (
	"fmt"
	"github.com/hypersleep/easyssh"
	"os/exec"
        "strings"
	"sync"
)

type Command struct {
        cmd string
        args []string
}

func BuildCommand(name string, arg []string) *Command {
	command := &Command{
                cmd: name,
                args: append([]string{name}, arg...),
        }
        return command; 
}

func (c *Command) getArgs() string {
        if len(c.args) > 0 {
               return strings.Join(c.args, " ")
        }
        return c.cmd
}

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
                statuses := make([]Status, s.LenCommands())
		for _, command := range s.commands {
			wg.Add(1)
			go func() {
				defer wg.Done()
				response, err := ssh.Run(command.getArgs())
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
