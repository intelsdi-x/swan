// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

var (
	currentUser, _ = user.Current()

	sshUserFlag        = conf.NewStringFlag("remote_ssh_login", "Login used for connecting to remote nodes. ", currentUser.Name)
	sshUserKeyPathFlag = conf.NewStringFlag("remote_ssh_key_path", fmt.Sprintf("Key for %q used for connecting to remote nodes.", sshUserFlag.Name), path.Join(currentUser.HomeDir, ".ssh/id_rsa"))

	sshPortFlag = conf.NewIntFlag("remote_ssh_port", "Port used for SSH connection to remote nodes. ", 22)
)

// RemoteConfig is configuration for Remote Executor.
type RemoteConfig struct {
	User    string
	KeyPath string

	Port int
}

// DefaultRemoteConfig returns default Remote Executor configuration from flags.
func DefaultRemoteConfig() RemoteConfig {
	return RemoteConfig{
		User:    sshUserFlag.Value(),
		KeyPath: sshUserKeyPathFlag.Value(),
		Port:    sshPortFlag.Value(),
	}
}

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	clientConfig *ssh.ClientConfig
	config       RemoteConfig
	targetHost   string

	// Note that by default on Decorate PID isolation is added at the end.
	commandDecorators isolation.Decorators
}

// NewRemoteFromIP returns a remote executo instance.
func NewRemoteFromIP(address string) (Executor, error) {
	return NewRemote(address, DefaultRemoteConfig())
}

// NewRemote returns a remote executor instance.
func NewRemote(address string, config RemoteConfig) (Executor, error) {
	return NewRemoteIsolated(address, config, isolation.Decorators{})
}

// NewRemoteIsolated returns a remote executor instance.
func NewRemoteIsolated(address string, config RemoteConfig, decorators isolation.Decorators) (Executor, error) {
	authMethod, err := getAuthMethod(config.KeyPath)
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	return Remote{
		targetHost:        address,
		config:            config,
		clientConfig:      clientConfig,
		commandDecorators: decorators,
	}, nil
}

// getAuthMethod which uses given key.
func getAuthMethod(keyPath string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "reading key %q failed", keyPath)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing private key %q failed", keyPath)
	}

	return ssh.PublicKeys(key), nil
}

// Name returns User-friendly name of executor.
func (remote Remote) Name() string {
	return "Remote Executor"
}

// Execute runs the command given as input.
// Returned Task Handle is able to stop & monitor the provisioned process.
func (remote Remote) Execute(command string) (TaskHandle, error) {
	connectionCommand := fmt.Sprintf("%s:%d", remote.targetHost, remote.config.Port)
	connection, err := ssh.Dial(
		"tcp",
		connectionCommand,
		remote.clientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "ssh.Dial to '%s@%s' for command %q failed",
			remote.clientConfig.User, connectionCommand, command)
	}

	session, err := newSessionWithPty(connection)
	if err != nil {
		return nil, errors.Wrapf(err, "connection.sewSessionWithPty for command %q failed with error %v", command, err)
	}

	output, err := createOutputDirectory(command, "remote")
	if err != nil {
		return nil, errors.Wrapf(err, "createOutputDirectory for command %q failed", command)
	}
	stdoutFile, stderrFile, err := createExecutorOutputFiles(output)
	if err != nil {
		removeDirectory(output)
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

	stringForSh = fmt.Sprintf("%s", stringForSh)

	log.Debug("Starting '", stringForSh, "' remotely on '", remote.targetHost, "'")
	err = session.Start(stringForSh)
	if err != nil {
		return nil, errors.Wrapf(err, "session.Start for command %q failed", command)
	}

	log.Debug("Started remote command")

	// hasProcessExited channel is closed when launched process exits.
	hasProcessExited := make(chan struct{})

	// TODO(bplotka): Move exit code constants to global executor scope.
	const successExitCode = int(0)
	const errorExitCode = int(-1)

	exitCodeInt := errorExitCode
	var exitCode *int
	exitCode = &exitCodeInt

	taskHandle := remoteTaskHandle{
		session:          session,
		connection:       connection,
		command:          command,
		stdoutFilePath:   stdoutFile.Name(),
		stderrFilePath:   stderrFile.Name(),
		host:             remote.targetHost,
		exitCode:         exitCode,
		hasProcessExited: hasProcessExited,
	}

	// Wait for remote task in go routine.
	go func() {
		defer func() {
			session.Close()
			connection.Close()
		}()
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
		close(hasProcessExited)

		err = syncAndClose(stdoutFile)
		if err != nil {
			log.Errorf("Cannot syncAndClose stdout file: %s", err.Error())
		}
		err = syncAndClose(stderrFile)
		if err != nil {
			log.Errorf("Cannot syncAndClose stderrFile file: %s", err.Error())
		}
	}()

	// Best effort potential way to check if binary is started properly.
	taskHandle.Wait(100 * time.Millisecond)
	err = checkIfProcessFailedToExecute(command, remote.Name(), &taskHandle)
	if err != nil {
		return nil, err
	}
	return &taskHandle, nil
}

// Final wait for the command to exit
const killTimeout = 5 * time.Second

// Period between sending SIGTERM  and SIGKILL
const killWaitTimeout = 100 * time.Millisecond

// remoteTaskHandle implements TaskHandle interface.
type remoteTaskHandle struct {
	session        *ssh.Session
	connection     *ssh.Client
	stdoutFilePath string
	stderrFilePath string
	host           string
	exitCode       *int

	// Command requested by User. This is how this TaskHandle presents.
	command string

	// This channel is closed immediately when process exits.
	// It is used to signal task termination.
	hasProcessExited chan struct{}
}

// isTerminated checks if channel processHasExited is closed. If it is closed, it means
// that wait ended and task is in terminated state.
func (taskHandle *remoteTaskHandle) isTerminated() bool {
	select {
	case <-taskHandle.hasProcessExited:
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

	err := taskHandle.session.Close()
	if err != nil {
		return errors.Wrapf(err, "could not close ssh session")
	}
	isTerminated := taskHandle.Wait(killWaitTimeout)
	if !isTerminated {
		return errors.New("cannot stop ssh session")
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
	return openFile(taskHandle.stdoutFilePath)
}

// StderrFile returns a file handle for file to the task's stderr file.
func (taskHandle *remoteTaskHandle) StderrFile() (*os.File, error) {
	return openFile(taskHandle.stderrFilePath)
}

// EraseOutput deletes the directory where stdout file resides.
func (taskHandle *remoteTaskHandle) EraseOutput() error {
	outputDir := filepath.Dir(taskHandle.stdoutFilePath)
	return removeDirectory(outputDir)
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
	case <-taskHandle.hasProcessExited:
		// If waitEndChannel is closed then task is terminated.
		return true
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		return false
	}
}

func (taskHandle *remoteTaskHandle) Name() string {
	return fmt.Sprintf("Remote %q on %q", taskHandle.command, taskHandle.Address())
}

func (taskHandle *remoteTaskHandle) Address() string {
	return taskHandle.host
}

// Killing the remote process related helper functions.
func newSessionWithPty(connection *ssh.Client) (*ssh.Session, error) {
	session, err := connection.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "newSessionWithPty: connection.NewSession failed")
	}

	terminal := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 80, 40, terminal)
	if err != nil {
		session.Close()
		return nil, errors.Wrapf(err, "newSessionWithPty: session.RequestPty failed")
	}
	return session, nil
}
