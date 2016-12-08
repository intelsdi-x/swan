package command

import (
	"bytes"
	"strings"
)

type Command struct {
	cmd string
	args []string
}

func NewCommand(name string, args []string)*Command{
	return &Command{
		name,
		args,
	}
}

func(command *Command) getCommand() string{
	var commandBuff bytes.Buffer
	commandBuff.WriteString(command.cmd)
	if len(command.args)>0 {
		commandBuff.WriteString(strings.Join(command.args, ""))
	}
	return commandBuff.String()
}
