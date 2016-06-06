package executor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/workloads"
	. "github.com/smartystreets/goconvey/convey"
	"time"
)

const (
	EnvHost          = "REMOTE_EXECUTOR_TEST_HOST"
	EnvUser          = "REMOTE_EXECUTOR_USER"
	EnvKey           = "REMOTE_EXECUTOR_SSH_KEY"
	EnvMemcachedPath = "REMOTE_EXECUTOR_MEMCACHED_BIN_PATH"
	EnvMemcachedUser = "REMOTE_EXECUTOR_MEMCACHED_USER"
	EnvCmd           = "REMOTE_EXECUTOR_CMD"
)

var terminal ssh.TerminalModes

func init() {
	terminal = ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
}

func isMemcachedEnvironmentReady() bool {
	if os.Getenv(EnvHost) == "" {
		return false
	}
	if os.Getenv(EnvUser) == "" {
		return false
	}
	if os.Getenv(EnvKey) == "" {
		return false
	}
	if os.Getenv(EnvMemcachedPath) == "" {
		return false
	}
	if os.Getenv(EnvMemcachedUser) == "" {
		return false
	}

	return true
}

func TestMemcachedRemoteProcessPidIsolation(t *testing.T) {
	if isMemcachedEnvironmentReady() {
		Convey("When I create remote executor for memcached", t, testMemcachedRemoteProcessPidIsolation)
	} else {
		SkipConvey("When I create remote executor for memcached", t, testMemcachedRemoteProcessPidIsolation)
	}
}

func isEnvironmentReady() bool {
	if os.Getenv(EnvHost) == "" {
		return false
	}
	if os.Getenv(EnvUser) == "" {
		return false
	}
	if os.Getenv(EnvKey) == "" {
		return false
	}
	if os.Getenv(EnvCmd) == "" {
		return false
	}

	return true
}

func TestCustomRemoteProcess(t *testing.T) {
	if isEnvironmentReady() {
		Convey("When I create remote executor", t, testCustomRemoteProcess)
	} else {
		SkipConvey("When I create remote executor", t, testCustomRemoteProcess)
	}
}

func testMemcachedRemoteProcessPidIsolation() {
	clientConfig, cErr := executor.NewClientConfig(os.Getenv(EnvUser), os.Getenv(EnvKey))
	So(cErr, ShouldBeNil)
	sshConfig := executor.NewSSHConfig(clientConfig, os.Getenv(EnvHost), 22)
	launcher := newMultipleMemcached(*sshConfig)

	Convey("I should be able to execute remote command and see the processes running", func() {
		task, err := launcher.Launch()

		client, err := ssh.Dial("tcp", os.Getenv(EnvHost)+":22", clientConfig)
		So(err, ShouldBeNil)
		defer client.Close()
		pids := soProcessesAreRunning(client, "memcached", 2)

		Convey("I should be able to stop remote task and all the processes should be terminated",
			func() {
				err := task.Stop()
				So(err, ShouldBeNil)
				_, err = task.ExitCode()
				So(err, ShouldBeNil)
				soProcessIsNotRunning(client, pids[0])
				soProcessIsNotRunning(client, pids[1])
			})
	})

}

func newMultipleMemcached(sshConfig executor.SSHConfig) workloads.Launcher {
	decors := isolation.Decorators{}
	unshare, _ := isolation.NewNamespace(syscall.CLONE_NEWPID)
	decors = append(decors, unshare)
	exec := executor.NewRemote(sshConfig, decors)

	return multipleMemcached{exec}
}

type multipleMemcached struct {
	executor executor.Executor
}

func (m multipleMemcached) Name() string {
	return "remote memcached"
}

func (m multipleMemcached) Launch() (executor.TaskHandle, error) {
	bin := os.Getenv(EnvMemcachedPath)
	username := os.Getenv(EnvMemcachedUser)
	return m.executor.Execute(
		fmt.Sprintf("/bin/bash -c \"%s -u %s -d && %s -u %s -p 54321\"", bin, username, bin, username))
}

func soProcessesAreRunning(client *ssh.Client, processName string, noOfPids int) (pids []string) {
	session, err := client.NewSession()
	So(err, ShouldBeNil)
	defer session.Close()
	err = session.RequestPty("xterm", 80, 40, terminal)
	So(err, ShouldBeNil)

	output, err := session.Output("pgrep " + processName)
	So(err, ShouldBeNil)
	pids = strings.Split(strings.Trim(string(output), "\n\r"), "\n")
	So(pids, ShouldHaveLength, noOfPids)

	for k, pid := range pids {
		pid = strings.Trim(pid, "\n\r")
		_, err = strconv.Atoi(pid)
		So(err, ShouldBeNil)
		pids[k] = pid

	}

	return pids
}

func soProcessIsNotRunning(client *ssh.Client, pid string) {
	session, err := client.NewSession()
	So(err, ShouldBeNil)
	defer session.Close()
	err = session.RequestPty("xterm", 80, 40, terminal)
	So(err, ShouldBeNil)
	_, err = session.Output("sudo cat /proc/" + pid + "/cmdline")
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldStartWith, "Process exited with: 1")

}

func testCustomRemoteProcess() {
	clientConfig, cErr := executor.NewClientConfig(os.Getenv(EnvUser), os.Getenv(EnvKey))
	So(cErr, ShouldBeNil)
	sshConfig := executor.NewSSHConfig(clientConfig, os.Getenv(EnvHost), 22)

	unshare, _ := isolation.NewNamespace(syscall.CLONE_NEWPID)
	remote := executor.NewRemote(*sshConfig, isolation.Decorators{unshare})

	Convey("I should be able to execute remote command and see the processes running", func() {
		task, err := remote.Execute(os.Getenv(EnvCmd))
		So(err, ShouldBeNil)
		defer func() {
			err := task.Stop()
			So(err, ShouldBeNil)
			_, err = task.ExitCode()
			So(err, ShouldBeNil)
			task.Clean()
			task.EraseOutput()
		}()

		task.Wait(500 * time.Millisecond)
	})

}
