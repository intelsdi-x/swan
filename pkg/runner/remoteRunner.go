package runner

import (
	"swan/pkg/command"
	"swan/pkg/sshConfig"
)

type RemoteRunner struct{
	sshConfig sshConfig.SshConfig
}

func NewRemoteRunner(sshConfig sshConfig.SshConfig) *RemoteRunner {
	return &RemoteRunner{
		sshConfig,
	}
}

func (remoteRunner *RemoteRunner) run(command command.Command) *Task{

	return 
}
