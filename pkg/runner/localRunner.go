package runner

import (
	"swan/pkg/isolation"
	"swan/pkg/command"
)

type LocalRunner struct{
	isolation isolation.Isolation
}

func NewLocalRunner(isolation isolation.Isolation) *LocalRunner {
	return &LocalRunner{
		isolation,
	}
}

func (localRunner *LocalRunner) run(command command.Command) *Task{

	return
}
