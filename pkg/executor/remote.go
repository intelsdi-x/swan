package executor

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"time"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig SSHConfig
}

// NewRemote returns a Local instance.
func NewRemote(sshConfig SSHConfig) *Remote {
	return &Remote{
		sshConfig,
	}
}

// Execute runs the command given as input.
// Returned Task pointer is able to stop & monitor the provisioned process.
func (remote Remote) Execute(command string) (Task, error) {
	log.Debug("Starting remotely", command)

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

	// Create temporary output files.
	currDir, _ := os.Getwd()
	stdoutFile, err := ioutil.TempFile(currDir, "swan_remote_executor_stdout_")
	if err != nil {
		return nil, err
	}
	stderrFile, err := ioutil.TempFile(currDir, "swan_remote_executor_stderr_")
	if err != nil {
		return nil, err
	}

	log.Debug("Created temporary files: ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	session.Stdout = stdoutFile
	session.Stderr = stderrFile

	err = session.Start(command)
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

	return newRemoteTask(session, stdoutFile, stderrFile, waitEndChannel, exitCode), nil
}

const killTimeout = 5 * time.Second

// remoteTask implements Task interface.
type remoteTask struct {
	session        *ssh.Session
	stdoutFile     *os.File
	stderrFile     *os.File
	waitEndChannel chan struct{}
	exitCode       *int
}

// newRemoteTask returns a RemoteTask instance.
func newRemoteTask(session *ssh.Session, stdoutFile *os.File, stderrFile *os.File,
	waitEndChannel chan struct{}, exitCode *int) *remoteTask {
	return &remoteTask{
		session:        session,
		stdoutFile:     stdoutFile,
		stderrFile:     stderrFile,
		waitEndChannel: waitEndChannel,
		exitCode:       exitCode,
	}
}

// isTerminated checks if waitEndChannel is closed. If it is closed, it means
// that wait ended and task is in terminated state.
func (task *remoteTask) isTerminated() bool {
	select {
	case <-task.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	default:
		return false
	}
}

// Stop terminates the remote task.
func (task *remoteTask) Stop() error {
	if task.isTerminated() {
		return nil
	}

	// Kill session.
	// NOTE: We need to find here a better way to stop task, since
	// closing channel just close the ssh session and some processes can be still running.
	// Some other approaches:
	// - sending Ctrl+C (very time based and not working currently)
	// - session.Signal does not work.
	// - gathering PID & killing the pid in separate session
	err := task.session.Close()
	if err != nil {
		log.Error(err)
		return err
	}

	// Checking if kill was successful.
	isTerminated := task.Wait(killTimeout)
	if !isTerminated {
		log.Error("Cannot terminate task")
		return errors.New("Cannot terminate task")

	}
	// No error, task terminated.
	return nil
}

// GetStatus returns a state of the task.
func (task *remoteTask) GetStatus() TaskState {
	if !task.isTerminated() {
		return RUNNING
	}

	return TERMINATED
}

// GetExitCode returns a exitCode. If task is not terminated it returns error.
func (task *remoteTask) GetExitCode() (int, error) {
	if !task.isTerminated() {
		return -1, errors.New("Task is not terminated")
	}

	return *task.exitCode, nil
}

// GetStdoutFile returns a file handle for file to the task's stdout file.
func (task *remoteTask) GetStdoutFile() (*os.File, error) {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return nil, err
	}

	task.stdoutFile.Seek(0, os.SEEK_SET)
	return task.stdoutFile, nil
}

// GetStderrFile returns a file handle for file to the task's stderr file.
func (task *remoteTask) GetStderrFile() (*os.File, error) {
	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return nil, err
	}

	task.stderrFile.Seek(0, os.SEEK_SET)
	return task.stderrFile, nil
}

// Clean removes files to which stdout and stderr of executed command was written.
func (task *remoteTask) Clean() error {
	// Close stdout.
	stdoutErr := task.stdoutFile.Close()

	// Close stderr.
	stderrErr := task.stderrFile.Close()

	if stdoutErr != nil {
		return stdoutErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (task *remoteTask) EraseOutput() error {
	// Remove stdout file.
	if err := os.Remove(task.stdoutFile.Name()); err != nil {
		return err
	}

	// Remove stderr file.
	if err := os.Remove(task.stderrFile.Name()); err != nil {
		return err
	}

	return nil
}

// Wait waits for the command to finish with the given timeout time.
// It returns true if task is terminated.
func (task *remoteTask) Wait(timeout time.Duration) bool {
	if task.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case <-task.waitEndChannel:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}
