package executor

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Remote provisioning is responsible for providing the execution environment
// on remote machine via ssh.
type Remote struct {
	sshConfig *SSHConfig
	// Note that by default on Decorate PID isolation is added at the end.
	commandDecorators isolation.Decorators
	// Unique ID for the command which will be searched on the remote host.
	unshareUUID string
}

// NewRemote returns a Remote instance.
func NewRemote(sshConfig *SSHConfig) Remote {
	var uuidStr string

	uuid, err := uuid.NewV4()
	if err != nil {
		uuidStr = string(time.Now().Unix())
	} else {
		uuidStr = uuid.String()
	}
	return Remote{
		sshConfig:         sshConfig,
		commandDecorators: []isolation.Decorator{},
		unshareUUID:       uuidStr,
	}
}

// NewRemoteIsolated returns a Remote instance.
func NewRemoteIsolated(sshConfig *SSHConfig, decorators isolation.Decorators) Remote {
	var uuidStr string
	uuid, err := uuid.NewV4()
	if err != nil {
		uuidStr = string(time.Now().Unix())
	} else {
		uuidStr = uuid.String()
	}

	return Remote{
		sshConfig:         sshConfig,
		commandDecorators: decorators,
		unshareUUID:       uuidStr,
	}
}

// Name returns user-friendly name of executor.
func (remote Remote) Name() string {
	return "Remote Executor"
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

	// Obligatory Pid namespace and a hint as comment. It will be carried to remote system.
	// On the server the example command will look the following:
	// unshare --fork --pid --mount-proc sh -c /opt/mutilate -A #d2857955-942c-4436-4d75-635640d2bbe5
	stringForSh = fmt.Sprintf(`unshare --fork --pid --mount-proc sh -c '%s #%s'`, stringForSh, remote.unshareUUID)

	log.Debug("Starting '", stringForSh, "' remotely")
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

	taskHandle := newRemoteTaskHandle(session, connection, stdoutFile.Name(), stderrFile.Name(),
		remote.sshConfig.Host, remote.unshareUUID, exitCode, hasProcessExited)

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

	return taskHandle, nil
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
	uuid           string
	exitCode       *int

	// This channel is closed immediately when process exits.
	// It is used to signal task termination.
	hasProcessExited chan struct{}
}

// newRemoteTaskHandle returns a remoteTaskHandle instance.
func newRemoteTaskHandle(
	session *ssh.Session,
	connection *ssh.Client,
	stdoutFilePath string,
	stderrFilePath string,
	host string,
	uuid string,
	exitCode *int,
	processHasExited chan struct{}) *remoteTaskHandle {
	return &remoteTaskHandle{
		session:          session,
		connection:       connection,
		stdoutFilePath:   stdoutFilePath,
		stderrFilePath:   stderrFilePath,
		host:             host,
		uuid:             uuid,
		exitCode:         exitCode,
		hasProcessExited: processHasExited,
	}
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
	err := killRemoteTaskWithID(taskHandle.connection, taskHandle.uuid, "SIGTERM")
	if err != nil {
		// Error here means that kill did not send signal.
		return errors.Wrapf(err, "remoteTaskHandle.Stop() failed to kill remote task with uuid %d at %s with signal SIGTERM", taskHandle.Address(), taskHandle.uuid)
	}

	// Wait for the Execute's go routine to update status.
	// If Wait exits with terminated status then there is no problem.
	// If Wait exits with not-terminated status then:
	//    a) task ignored SIGTERM
	//    b) task is killed but status has not changed yet - race here.
	isTerminated := taskHandle.Wait(killWaitTimeout)
	if !isTerminated {
		// Task is not terminated. Try kill it with SIGKILL.
		// Note that race can occur here so ignore errors.
		// Go routine may close session any time (defers).
		_ = killRemoteTaskWithID(taskHandle.connection, taskHandle.uuid, "SIGKILL")

		// Checking if kill was successful.
		isTerminated = taskHandle.Wait(killTimeout)
		if !isTerminated {
			return errors.Wrapf(err, "remoteTaskHandle.Stop() probably failed to kill remote task at %s with signal SIGKILL. Verify by 'ps aux | grep %s' on that host.", taskHandle.Address(), taskHandle.uuid)

		}
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

// Deprecated: Does nothing.
func (taskHandle *remoteTaskHandle) Clean() error {
	return nil
}

// EraseOutput removes task's stdout & stderr files.
func (taskHandle *remoteTaskHandle) EraseOutput() error {
	outputDir, _ := path.Split(taskHandle.stdoutFilePath)

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
	case <-taskHandle.hasProcessExited:
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

func getRemoteCmdOutput(connection *ssh.Client, cmd string) (string, error) {
	session, err := newSessionWithPty(connection)
	if err != nil {
		return "", errors.Wrapf(err, "getRemoteCmdOutput: newSessionWithPty failed")
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "getRemoteCmdOutput: session.Output failed for '%s'", cmd)
	}
	return strings.TrimSpace(string(output)), nil
}

func getUnshareProcessID(connection *ssh.Client, uuid string) (string, error) {
	// Get unshare process that in command line has also given uuid.
	// 1. ps -o pid -o cmd ax
	//    - prints only PID and CMD collumns. 'a' - all process,
	//      'x' - even if they are not attached to terminal.
	// Example:
	// [root@localhost]$ ps -o pid -o cmd ax
	//   PID CMD
	//     1 /usr/lib/systemd/systemd --switched-root --system --deserialize 21
	//     2 [kthreadd]
	//     3 [ksoftirqd/0]
	//   ...
	// 23403 sshd: root@notty
	// 23406 unshare --fork --pid --mount-proc sh -c /home/vagrant/.../mutilate -A #d2857955-942c-4436-4d75-635640d2bbe5
	// 23411 /home/vagrant/.../mutilate -A
	//
	// 2. Grep for 'unshare'. '[e]' prevents grep to find itself in proceess list.
	//
	// 3. Second grep searches for given uuid in all found 'unshare' process.
	// Note that there could be more that one 'unshare' that's why uuid is
	// added as a comment to command.
	cmd := fmt.Sprintf(`ps -o pid -o cmd ax | grep unshar[e] | grep -e %q`, uuid)
	// Grep returns 0 - success, 1 - pattern not found, 2 - error.
	output, err := getRemoteCmdOutput(connection, cmd)
	if err != nil {
		return "", errors.Wrapf(err, "getUnshareProcessID getRemoteCmdOutput failed for command '%s'", cmd)
	}
	// Output from search is '<pid> <full command>'.
	unsharePid := strings.Split(output, " ")[0]
	return unsharePid, nil
}

func getPidNamespaceInit(connection *ssh.Client, unsharePid string) (string, error) {
	// Find process to which 'unsharePid' is a parent. Print only found pit.
	cmd := "ps -opid= --ppid " + unsharePid
	output, err := getRemoteCmdOutput(connection, cmd)
	if err != nil {
		return "", errors.Wrapf(err, "getPidNamespaceInit getRemoteCmdOutput failed for command '%s'", cmd)
	}
	childPid := strings.TrimSpace(output)
	return childPid, nil
}

func killRemotePid(connection *ssh.Client, sig string, pid string) error {
	session, err := newSessionWithPty(connection)
	if err != nil {
		return errors.Wrapf(err, "killRemotePid newSessionWithPty failed.")
	}
	defer session.Close()
	err = session.Run(fmt.Sprintf("kill -%s %s", sig, pid))
	// Kill return 'success' if signal was sent
	return err
}

func killRemoteTaskWithID(connection *ssh.Client, uuid string, signal string) error {
	// 1. Find 'unshare' process which has 'uuid' in command line attached. Return pid of that 'unshare'.
	unsharePid, err := getUnshareProcessID(connection, uuid)
	if err != nil {
		return errors.Wrapf(err, "killRemoteTaskWithID: getUnshareProcessID failed for uuid '%s'", uuid)
	}
	// 2. Find 'unshare' child - this will be init process in the PID namespace and killing it
	//    will result in killing all processes in that namespace.
	initPid, err := getPidNamespaceInit(connection, unsharePid)
	if err != nil {
		return errors.Wrapf(err, "killRemoteTaskWithID: getPidNamespaceInit failed for unsharePid '%s'", unsharePid)
	}
	// 3. Send kill signal to unshare's child - namespace init process.
	err = killRemotePid(connection, signal, initPid)
	if err != nil {
		return errors.Wrapf(err, "killRemoteTaskWithID: failed to send signal %s to process '%s'", signal, initPid)
	}
	return err
}
