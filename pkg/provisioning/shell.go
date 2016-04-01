package provisioning

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hypersleep/easyssh"
	"os/exec"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

// Shell is responsible for providing the execution environment via bash or bash & ssh
// in case of remote execution. It also needed to setup given isolation
// using Isolation Manager.
type Shell struct{}

// NewShell returns a Shell instance.
func NewShell() Shell {
	return Shell{}
}

func targetIsLocal(targetHost string) bool {
	return targetHost == "local" || targetHost == "localhost" || targetHost == "127.0.0.1"
}


// Execute runs the task given in parameters
// parallel. Returned channel specifies when task has completed or failed.
func (s Shell) Execute(command string, targetHost string,
					   isolations []isolation.Isolation) (<-chan Status) {
	statusCh := make(chan Status)

	if targetIsLocal(targetHost) {
		taskPidCh := make(chan int)
		taskPid := -1

		// TODO(bplotka): We can move isolation to both ways of deploy (remote & local)
		// when we find a way to get PID from process under ssh session.
		// Do initialization of the isolation synchronously.
		for _, isolation := range isolations {
			isolation.Init(targetHost)
		}

		// Run task in shell locally.
		go func() {
			log.Debug("Starting ", command)
			cmd := exec.Command("sh", "-c", command)
			cmd.Start()

			log.Debug("Started with pid ", cmd.Process.Pid)
			// Report the process id.
			taskPidCh <- cmd.Process.Pid
			// Wait for task completion.
			cmd.Wait()
			log.Debug("Ended ", command)

			// TODO(bplotka): Fetch status code.

			output, _ := cmd.Output()
			statusCh <- Status{0, string(output)}
		}()

		// Get PID
		taskPid = <-taskPidCh

		// Perform rest of the isolation synchronously.
		for _, isolation := range isolations {
			isolation.Perform(taskPid)
		}
	} else {
		// Run task in shell remotelly via ssh.
		ssh := &easyssh.MakeConfig{
			User:   "root",
			Server: "localhost",
			Key:    "/.ssh/id_rsa",
			Port:   "22",
		}
		go func() {
			log.Debug("Starting remote ", command)

			// TODO(bplotka): Find a way to have a PID here for future isolation.
			response, err := ssh.Run(command)

			if err != nil {
				panic("Can't run remote command: " + err.Error())
			}

			log.Debug("Ended ", command)

			// TODO(bplotka): Fetch status code.
			statusCh <- Status{0, response}
		}()
	}

	return statusCh
}
