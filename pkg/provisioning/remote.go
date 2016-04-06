package provisioning

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hypersleep/easyssh"
)

// RemoteTask implements Task interface.
type RemoteTask struct{
	// TODO(bp): Alicja to fill that.
}

// NewRemoteTask returns a LocalTask instance.
func NewRemoteTask() *RemoteTask {
	t := &RemoteTask{
	}
	return t
}

// Stop terminates the remote task.
func (task *RemoteTask) Stop() {
	// TODO(bp): Stop pid with.
	// TODO(bp): Alicja to fill that.
	panic("Not implemented")
}

// Status gets status of the remote task.
func (task RemoteTask) Status() Status {
	// TODO(bp): Get status.
	// TODO(bp): Alicja to fill that.
	return Status{}
}

// Wait blocks until process is terminated or timeout appeared.
func (task *RemoteTask) Wait(timeoutSeconds int) bool {
	// TODO(bp): Alicja to fill that.
	panic("Not implemented")

	return true
}

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct{
	// TODO(bp): Alicja to fill that.
}

// NewRemote returns a Local instance.
func NewRemote() Remote {
	l := Remote{
	}
	return l
}


// Run runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (l Remote) Run(command string) (Task, error) {
	statusCh := make(chan Status)

	// Run task in shell remotely via ssh.
	ssh := &easyssh.MakeConfig{
	User:   "root",
	Server: "localhost",
	Key:    "/.ssh/id_rsa",
	Port:   "22",
	}
	go func() {
		log.Debug("Starting remote ", command)

		response, err := ssh.Run(command)

		if err != nil {
			panic("Can't run remote command: " + err.Error())
		}

		log.Debug("Ended ", command)

		// TODO(bplotka): Fetch status code.
		statusCh <- Status{0, response}
	}()

	// TODO(bp): Alicja to fill that.
	//t := NewRemoteTask(statusCh)

	return NewRemoteTask(), nil
}
