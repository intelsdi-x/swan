package executor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"golang.org/x/crypto/ssh"
	"os/user"
	"syscall"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig         SSHConfig
	commandDecorators isolation.Decorator
}

// NewRemote returns a Remote instance.
func NewRemote(sshConfig SSHConfig, decorator isolation.Decorator) *Remote {
	return &Remote{
		sshConfig,
		decorator,
	}
}

// NewRemote returns a Remote instance with default SSH configuration for specified host and
// PID namespace enabled.
func NewRemoteWithDefaultConfig(host string) (*Remote, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")
	if err != nil {
		return nil, err
	}

	sshConfig := NewSSHConfig(clientConfig, host, 22)
	isolationPid, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
	if err != nil {
		return nil, err
	}

	return &Remote{
		sshConfig:         *sshConfig,
		commandDecorators: isolationPid,
	}, nil
}

// Execute runs the command given as input.
// Returned Task Handle is able to stop & monitor the provisioned process.
func (remote Remote) Execute(command string) (TaskHandle, error) {
	log.Debug("Starting ", command, "remotely")

	connection, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", remote.sshConfig.host, remote.sshConfig.port),
		remote.sshConfig.clientConfig)
	if err != nil {
		return nil, err
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, err
	}

	terminal := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 80, 40, terminal)
	if err != nil {
		session.Close()
		return nil, err
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(command, "remote")
	if err != nil {
		return nil, err
	}

	log.Debug("Created temporary files: ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	session.Stdout = stdoutFile
	session.Stderr = stderrFile

	err = session.Start(remote.commandDecorators.Decorate(command))
	if err != nil {
		return nil, err
	}

	log.Debug("Started remote command")

	// Wait End channel is for checking the status of the Wait. If this channel is closed,
	// it means that the wait is completed (either with error or not)
	// This channel will not be used for passing any message.
	waitEndChannel := make(chan struct{})

	// TODO(bplotka): Move exit code constants to global executor scope.
	const successExitCode = int(0)
	const errorExitCode = int(-1)

	exitCodeInt := errorExitCode
	var exitCode *int
	exitCode = &exitCodeInt

	// Wait for local task in go routine.
	go func() {
		defer close(waitEndChannel)
		defer session.Close()

		*exitCode = successExitCode
		// Wait for task completion.
		err := session.Wait()
		if err != nil {
			if exitError, ok := err.(*ssh.ExitError); !ok {
				// In case of NON Exit Errors we are not sure if task does
				// terminate so panic.
				log.Panic("Waiting for remote task failed. ", err)
			} else {
				*exitCode = exitError.Waitmsg.ExitStatus()
			}
		}

		log.Debug(
			"Ended ", command,
			" with output in file: ", stdoutFile.Name(),
			" with err output in file: ", stderrFile.Name(),
			" with status code: ", *exitCode)
	}()

	return newRemoteTaskHandle(session, stdoutFile, stderrFile,
		remote.sshConfig.host, waitEndChannel, exitCode), nil
}

const killTimeout = 5 * time.Second

// remoteTaskHandle implements TaskHandle interface.
type remoteTaskHandle struct {
	session        *ssh.Session
	stdoutFile     *os.File
	stderrFile     *os.File
	host           string
	waitEndChannel chan struct{}
	exitCode       *int
}

// newRemoteTaskHandle returns a remoteTaskHandle instance.
func newRemoteTaskHandle(session *ssh.Session, stdoutFile *os.File, stderrFile *os.File,
	host string, waitEndChannel chan struct{}, exitCode *int) *remoteTaskHandle {
	return &remoteTaskHandle{
		session:        session,
		stdoutFile:     stdoutFile,
		stderrFile:     stderrFile,
		host:           host,
		waitEndChannel: waitEndChannel,
		exitCode:       exitCode,
	}
}

// isTerminated checks if waitEndChannel is closed. If it is closed, it means
// that wait ended and task is in terminated state.
func (taskHandle *remoteTaskHandle) isTerminated() bool {
	select {
	case <-taskHandle.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	default:
		return false
	}
}

// Stop terminates the remote task.
func (taskHandle *remoteTaskHandle) Stop() error {
	if taskHandle.isTerminated() {
		return nil
	}

	// Kill session.
	// NOTE: We need to find here a better way to stop task, since
	// closing channel just close the ssh session and some processes can be still running.
	// Some other approaches:
	// - sending Ctrl+C (very time based and not working currently)
	// - session.Signal does not work.
	// - gathering PID & killing the pid in separate session
	err := taskHandle.session.Close()
	if err != nil {
		log.Error(err)
		return err
	}

	// Checking if kill was successful.
	isTerminated := taskHandle.Wait(killTimeout)
	if !isTerminated {
		log.Error("Cannot terminate task")
		return errors.New("Cannot terminate task")

	}
	// No error, task terminated.
	return nil
}

// Status returns a state of the task.
func (taskHandle *remoteTaskHandle) Status() TaskState {
	if !taskHandle.isTerminated() {
		return RUNNING
	}

	return TERMINATED
}

// ExitCode returns a exitCode. If task is not terminated it returns error.
func (taskHandle *remoteTaskHandle) ExitCode() (int, error) {
	if !taskHandle.isTerminated() {
		return -1, errors.New("Task is not terminated")
	}

	return *taskHandle.exitCode, nil
}

// StdoutFile returns a file handle for file to the task's stdout file.
func (taskHandle *remoteTaskHandle) StdoutFile() (*os.File, error) {
	if _, err := os.Stat(taskHandle.stdoutFile.Name()); err != nil {
		return nil, err
	}

	taskHandle.stdoutFile.Seek(0, os.SEEK_SET)
	return taskHandle.stdoutFile, nil
}

// StderrFile returns a file handle for file to the task's stderr file.
func (taskHandle *remoteTaskHandle) StderrFile() (*os.File, error) {
	if _, err := os.Stat(taskHandle.stderrFile.Name()); err != nil {
		return nil, err
	}

	taskHandle.stderrFile.Seek(0, os.SEEK_SET)
	return taskHandle.stderrFile, nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (taskHandle *remoteTaskHandle) Clean() error {
	// Close stdout.
	stdoutErr := taskHandle.stdoutFile.Close()

	// Close stderr.
	stderrErr := taskHandle.stderrFile.Close()

	if stdoutErr != nil {
		return stdoutErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (taskHandle *remoteTaskHandle) EraseOutput() error {
	outputDir, _ := path.Split(taskHandle.stdoutFile.Name())

	// Remove temporary directory created for stdout and stderr.
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}
	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (taskHandle *remoteTaskHandle) Wait(timeout time.Duration) bool {
	if taskHandle.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case <-taskHandle.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}

func (taskHandle *remoteTaskHandle) Address() string {
	return taskHandle.host
}
