package executor

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig         *SSHConfig
	commandDecorators isolation.Decorators
}

// NewRemote returns a Remote instance.
func NewRemote(sshConfig *SSHConfig) Remote {
	return Remote{
		sshConfig:         sshConfig,
		commandDecorators: []isolation.Decorator{},
	}
}

// NewRemoteIsolated returns a Remote instance.
func NewRemoteIsolated(sshConfig *SSHConfig, decorators isolation.Decorators) Remote {
	return Remote{
		sshConfig:         sshConfig,
		commandDecorators: decorators,
	}
}

// Execute runs the command given as input.
// Returned Task Handle is able to stop & monitor the provisioned process.
func (remote Remote) Execute(command string) (TaskHandle, error) {
	connectionCommand := fmt.Sprintf("%s:%d", remote.sshConfig.Host, remote.sshConfig.Port)
	connection, err := ssh.Dial(
		"tcp",
		connectionCommand,
		remote.sshConfig.ClientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "ssh.Dial to '%s@%s' for command %q failed",
			remote.sshConfig.ClientConfig.User, connectionCommand, command)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "connection.NewSession for command %q failed", command)
	}

	terminal := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 80, 40, terminal)
	if err != nil {
		session.Close()
		return nil, errors.Wrapf(err, "session.RequestPty for command %q failed", command)
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(command, "remote")
	if err != nil {
		return nil, errors.Wrapf(err, "createExecutorOutputFiles for command %q failed", command)
	}

	log.Debug("Created temporary files: ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	session.Stdout = stdoutFile
	session.Stderr = stderrFile

	// Escape the quotes characters for `sh -c`.
	stringForSh := remote.commandDecorators.Decorate(command)
	stringForSh = strings.Replace(stringForSh, "'", "\\'", -1)
	stringForSh = strings.Replace(stringForSh, "\"", "\\\"", -1)

	log.Debug("Starting '", stringForSh, "' remotely")

	// huponexit` ensures that the process will be killed when ssh connection will be closed.
	err = session.Start(fmt.Sprintf("shopt -s huponexit; sh -c %q", stringForSh))
	if err != nil {
		return nil, errors.Wrapf(err, "session.Start for command %q failed", command)
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
		defer session.Close() // Closing a session is not enough to close connection.
		defer connection.Close()

		*exitCode = successExitCode
		// Wait for task completion.
		err := session.Wait()
		if err != nil {
			if exitError, ok := err.(*ssh.ExitError); !ok {
				// In case of NON Exit Errors we are not sure if task does
				// terminate so panic.
				err = errors.Wrap(err, "wait returned with NON exit error")
				log.Panicf("Waiting for remote task failed %+v", err)
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
		remote.sshConfig.Host, waitEndChannel, exitCode), nil
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
		return errors.Wrap(err, "session.Close failed")
	}

	// Checking if kill was successful.
	isTerminated := taskHandle.Wait(killTimeout)
	if !isTerminated {
		return errors.New("cannot terminate task")

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
		return -1, errors.New("task is not terminated")
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
