package executor

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/parse"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/env"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	numAgents     = 3
	memcachedIP   = "10.141.141.10"
	masterIP      = "10.141.141.20"
	agent1IP      = "10.141.141.20"
	agent2IP      = "10.141.141.21"
	agent3IP      = "10.141.141.22"
	memcachedPort = 11212
)

func launchLocalMemcached(t *testing.T) executor.TaskHandle {

	localExecutor := executor.NewLocal()
	memcachedDefConfig := memcached.DefaultMemcachedConfig()
	memcachedDefConfig.User = env.GetOrDefault("USER", memcachedDefConfig.User)
	memcachedDefConfig.Port = memcachedPort
	memcachedDefConfig.NumThreads = 1

	memcachedLauncher := memcached.New(localExecutor, memcachedDefConfig)
	taskHandle, err := memcachedLauncher.Launch()
	if err != nil {
		t.Fatal("Cannot start local memcached instance: " + err.Error())
	}
	return taskHandle
}

func cleanMemcached(taskHandle executor.TaskHandle, t *testing.T) error {
	err := taskHandle.Stop()
	if err != nil {
		t.Fatal("Failed to stop local memcached instalce: " + err.Error())
	}
	taskHandle.Wait(0)
	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		t.Fatal("Failed to get local memcached exit code: " + err.Error())
	}
	if exitCode != -1 {
		t.Fatalf("Local memcached stopped incorrectly! Exit code: %d", exitCode)
	}
	err = taskHandle.EraseOutput()
	if err != nil {
		t.Fatal("Failed to clean output from local mutilate instance: " + err.Error())
	}
	return nil
}

func newSSHConfig(ip string) *executor.SSHConfig {
	return &executor.SSHConfig{
		ClientConfig: &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{ssh.Password("vagrant")},
		},
		Host: ip,
		Port: 22,
	}
}

func setupMutilate() executor.LoadGenerator {

	masterRemoteExecutor := executor.NewRemote(newSSHConfig(masterIP))
	agentsRemoteExecutors := []executor.Executor{}

	agentsRemoteExecutors = append(agentsRemoteExecutors, executor.NewRemote(newSSHConfig(agent1IP)))
	agentsRemoteExecutors = append(agentsRemoteExecutors, executor.NewRemote(newSSHConfig(agent2IP)))
	agentsRemoteExecutors = append(agentsRemoteExecutors, executor.NewRemote(newSSHConfig(agent3IP)))

	mutilateConfig := mutilate.DefaultMutilateConfig()

	mutilateConfig.TuningTime = 10 * time.Second
	// Ensure files are removed afterwards.
	mutilateConfig.ErasePopulateOutput = true
	mutilateConfig.EraseTuneOutput = true
	mutilateConfig.WarmupTime = 3 * time.Second
	// Note: not sure if custom percentile is working correctly.
	// TODO: added a custom percentile integration test.
	mutilateConfig.LatencyPercentile = "99"
	mutilateConfig.MemcachedHost = memcachedIP
	mutilateConfig.MemcachedPort = memcachedPort
	mutilateConfig.ErasePopulateOutput = true
	mutilateConfig.EraseTuneOutput = true

	mutilateConfig.AgentThreads = 1
	mutilateConfig.AgentConnections = 1
	mutilateConfig.AgentConnectionsDepth = 1

	mutilateConfig.MasterThreads = 1
	mutilateConfig.MasterConnections = 1
	mutilateConfig.MasterConnectionsDepth = 1

	mutilateLoadGenerator := mutilate.NewCluster(masterRemoteExecutor,
		agentsRemoteExecutors, mutilateConfig)
	return mutilateLoadGenerator
}

// getMemcachedStats helper read and parse "stats" memcached command and return map key -> value.
// https://github.com/memcached/memcached/blob/master/doc/protocol.txt#L511
func getLocalMemcachedStats(t *testing.T) (currItems, getCount int) {
	const (
		statsCmd         = "stats\n"
		mcStatsReplySize = 4096 // Enough size to get whole response from memcached.
	)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", memcachedPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if n, err := conn.Write([]byte(statsCmd)); err != nil || n != len(statsCmd) {
		t.Fatalf("couldn't write to memcached expected number err=%s of bytes=%d", err, n)
	}

	buf := make([]byte, mcStatsReplySize)
	if _, err = conn.Read(buf); err != nil {
		t.Fatal(err)
	}

	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "END") {
			break
		}
		var key, value string
		_, err := fmt.Sscanf(line, "STAT %s %s", &key, &value)
		if err != nil {
			t.Fatal(err)
		}
		switch key {
		case "curr_items":
			if currItems, err = strconv.Atoi(value); err != nil {
				t.Fatal(err)
			}
		case "cmd_get":
			if getCount, err = strconv.Atoi(value); err != nil {
				t.Fatal(err)
			}
		}
	}
	return currItems, getCount
}

func TestClusteredTaskHandle(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//log.SetOutput(os.Stderr)

	memcachedTaskHandle := launchLocalMemcached(t)

	defer cleanMemcached(memcachedTaskHandle, t)

	stopped := memcachedTaskHandle.Wait(2 * time.Second)
	if stopped {
		t.Fatal("Memcached not running! It should now.")
	}

	Convey("When launched new local memcached", t, func() {
		Convey("It should be running by now", func() {
			So(memcachedTaskHandle.Status(), ShouldEqual, executor.RUNNING)
		})

		Convey("It should have reset statistics", func() {
			currItem, getCount := getLocalMemcachedStats(t)
			So(currItem, ShouldEqual, 0)
			So(getCount, ShouldEqual, 0)
		})
	})

	mutilateLG := setupMutilate()

	Convey("With clustered mutilate configured", t, func() {
		Convey("It shall populate memcached", func() {
			err := mutilateLG.Populate()
			So(err, ShouldBeNil)
			currItems, _ := getLocalMemcachedStats(t)
			So(currItems, ShouldBeGreaterThan, 0)
		})
		Convey("It shall perform tune", func() {
			_, prevGetCount := getLocalMemcachedStats(t)
			achievedLoad, achievedSLI, err := mutilateLG.Tune(5000)
			So(err, ShouldBeNil)
			So(achievedLoad, ShouldNotEqual, 0)
			So(achievedSLI, ShouldNotEqual, 0)
			_, currentGetCount := getLocalMemcachedStats(t)
			So(currentGetCount, ShouldBeGreaterThan, prevGetCount)
		})
		Convey("It shall stress memcached", func() {
			_, prevGetCount := getLocalMemcachedStats(t)
			mutilateHandle, err := mutilateLG.Load(10, 10*time.Second)
			So(err, ShouldBeNil)
			mutilateHandle.Wait(0)
			So(mutilateHandle.Status(), ShouldEqual, executor.TERMINATED)
			exitcode, err := mutilateHandle.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode, ShouldEqual, 0)

			if err != nil {
				t.Fatalf("mutilate didn't stopped correctly err=%q", err)
			}

			out, err := mutilateHandle.StdoutFile()
			rawMetrics, err := parse.File(out.Name())

			SoNonZeroMetricExists := func(name string) {
				v, ok := rawMetrics.Raw[name]
				So(ok, ShouldBeTrue)
				So(v, ShouldBeGreaterThan, 0)
			}

			SoNonZeroMetricExists("qps")
			SoNonZeroMetricExists("avg")
			SoNonZeroMetricExists("std")
			SoNonZeroMetricExists("min")
			SoNonZeroMetricExists("percentile/99th")
			SoNonZeroMetricExists("percentile/custom")

			if err := mutilateHandle.EraseOutput(); err != nil {
				t.Fatal(err)

			}
			_, currGetCount := getLocalMemcachedStats(t)
			So(currGetCount, ShouldBeGreaterThan, prevGetCount)
		})
	})
}
