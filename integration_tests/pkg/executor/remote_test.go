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
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/ssh"
)

//const (
//	EnvHost          = "SWAN_REMOTE_EXECUTOR_TEST_HOST"
//	EnvUser          = "SWAN_REMOTE_EXECUTOR_USER"
//	EnvMemcachedPath = "SWAN_REMOTE_EXECUTOR_MEMCACHED_BIN_PATH"
//	EnvMemcachedUser = "SWAN_REMOTE_EXECUTOR_MEMCACHED_USER"
//)

// This tests required following setup:
// - id_rsa ssh keys in user home directory. [command ssh-keygen]
// - no password ssh session. [command ssh-copy-id localhost]
func TestRemote(t *testing.T) {
	Convey("Preparing Remote Executor to be tested on localhost", t, func() {

		remote, err := NewRemoteFromIP("127.0.0.1")
		if err != nil {
			t.Skip("Skipping remote executor test: " + err.Error())
		}

		Convey("And while using Remote Shell, the generic Executor test should pass", func() {
			testExecutor(t, remote)
		})
	})
}

//func TestRemoteProcessPidIsolation(t *testing.T) {
//	if isEnvironmentReady() {
//		Convey("When I create remote executor for memcached", t, testRemoteProcessPidIsolation)
//	} else {
//		SkipConvey("When I create remote executor for memcached", t, testRemoteProcessPidIsolation)
//	}
//}

func TestRemoteProcessPidIsolation(t *testing.T) {
	//user, err := user.Lookup(os.Getenv(EnvUser))
	//So(err, ShouldBeNil)

	//err = executor.ValidateSSHConfig(os.Getenv(EnvHost), user)
	//So(err, ShouldBeNil)

	//sshConfig, err := executor.NewSSHConfig(os.Getenv(EnvHost), 22, user)
	//So(err, ShouldBeNil)

	//launcher := newMultipleMemcached(*sshConfig)

	Convey("I should be able to execute remote command and see the processes running", func() {
		//task, err := launcher.Launch()
		//defer func() {
		//	task.Stop()
		//	task.EraseOutput()
		//}()

		//client, err := ssh.Dial("tcp", os.Getenv(EnvHost)+":22", sshConfig.ClientConfig)
		//So(err, ShouldBeNil)
		//defer client.Close()

		config := DefaultRemoteConfig()
		remote, err := NewRemote("127.0.0.1", config)
		if err != nil {
			t.Skip("Skipping remote executor test: " + err.Error())
		}

		user := config.User
		handle, err := remote.Execute(fmt.Sprintf("memcached -u %s -d && memcached -u %s -d", user, user))
		So(err, ShouldBeNil)

		pids := getPids(remote, "memcached", 2)

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

//func isEnvironmentReady() bool {
//	if value := os.Getenv(EnvHost); value == "" {
//		return false
//	}
//	if value := os.Getenv(EnvUser); value == "" {
//		return false
//	}
//
//	if value := os.Getenv(EnvMemcachedPath); value == "" {
//		return false
//	}
//	if value := os.Getenv(EnvMemcachedUser); value == "" {
//		return false
//	}
//
//	return true
//}

//var terminal ssh.TerminalModes
//
//func init() {
//	terminal = ssh.TerminalModes{
//		ssh.ECHO:          1,
//		ssh.TTY_OP_ISPEED: 14400,
//		ssh.TTY_OP_OSPEED: 14400,
//	}
//}

func getPids(remote Executor, processName string, noOfPids int) (pids []string) {
	//session, err := client.NewSession()
	//So(err, ShouldBeNil)
	//defer session.Close()
	//err = session.RequestPty("xterm", 80, 40, terminal)
	//So(err, ShouldBeNil)

	handle, err := remote.Execute("pgrep " + processName)
	So(err, ShouldBeNil)
	handle.Wait(0)

	exitCode, err := handle.ExitCode()
	So(err, ShouldBeNil)
	So(exitCode, ShouldEqual, 0)

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

//func newMultipleMemcached(sshConfig executor.SSHConfig) executor.Launcher {
//	decors := isolation.Decorators{}
//	unshare, _ := isolation.NewNamespace(syscall.CLONE_NEWPID)
//	decors = append(decors, unshare)
//	exec := executor.NewRemoteIsolated(&sshConfig, decors)
//
//	return multipleMemcached{exec}
//}
//
//type multipleMemcached struct {
//	executor executor.Executor
//}
//
//func (m multipleMemcached) Name() string {
//	return "remote memcached"
//}
//
//func (m multipleMemcached) Launch() (executor.TaskHandle, error) {
//	bin := os.Getenv(EnvMemcachedPath)
//	username := os.Getenv(EnvMemcachedUser)
//	return m.executor.Execute(
//		fmt.Sprintf("/bin/bash -c \"%s -u %s -d && %s -u %s -p 54321\"", bin, username, bin, username))
//}
//
