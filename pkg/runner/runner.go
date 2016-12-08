package runner

import (
	"swan/pkg/command"
)

type Runner interface{
	run(command command.Command) Task
}
