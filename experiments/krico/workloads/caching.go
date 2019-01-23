package workload

import (
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"strconv"
	kricosnapsession "github.com/intelsdi-x/swan/pkg/snap/sessions/krico"
	"github.com/intelsdi-x/backup/swan/experiments/krico/workloads"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
)

func CachingWorkload(){

	// Load OpenStack authentication variables from environment
	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	//
	// Workload
	//

	// Prepare configuration
	memcachedConfig := memcached.DefaultMemcachedConfig()

	// Prepare executor
	memcachedExecutorConfig := executor.DefaultOpenstackConfig(auth)
	memcachedExecutor := executor.NewOpenstack(&memcachedExecutorConfig)

	// Prepare launcher
	memcachedLauncher := memcached.New(memcachedExecutor, memcachedConfig)

	// Run workload
	memcachedWorkload, err := memcachedExecutor.Execute(memcachedLauncher.BuildCommand()+" -d")
	errutil.CheckWithContext(err, "Cannot launch Memcached!")

	// Stop workload on end
	defer memcachedWorkload.Stop()

	//
	// Aggressor
	//

	// Prepare configuration
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedWorkload.Address()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	//
	// Metrics collector (Snap)
	//

	// Start Snap on OpenStack host.

	startSnapExecutor, err := executor.NewRemoteFromIP(snapAddress.Value())
	errutil.CheckWithContext(err, "Cannot obtain Snap executor!")

	startSnapTaskHandle, err := startSnapExecutor.Execute(snapStartCommand)
	errutil.CheckWithContext(err, "Cannot execute start Snap command!")

	// Wait until start
	_, err = startSnapTaskHandle.Wait(0)
	errutil.CheckWithContext(err, "Cannot start Snap service!")

	// Todo: Check if it's needed
	// Wait until Snap will load plugins
	// time.Sleep(time.Second*10)

	//
	//	KRICO parameters
	//

	// Calculate workload parameters
	memory := float64(memcachedConfig.MaxMemoryMB / 1024) // total cache size [GiB]
	ratio, err := strconv.ParseFloat(mutilateConfig.Update, 64) // estimated get vs set ratio [0.0 - 1.0]
	errutil.CheckWithContext(err, "Cannot parse mutilate update argument (ratio parameter)!")
	clients := float64(mutilateConfig.MasterThreads * mutilateConfig.MasterConnections)

	//
	// Snap task
	//

	// Obtain workload cgroup
	cgroup, err := GetInstanceCgroup(memcachedExecutorConfig.Hypervisor.InstanceName, memcachedExecutorConfig.Hypervisor.Address)
	errutil.CheckWithContext(err, "Cannot obtain instance cgroup!")

	// Configure task
	kricoSnapTaskConfig := kricosnapsession.DefaultConfig(cgroup, memcachedExecutorConfig.Hypervisor.InstanceName)
	kricoSnapTaskConfig.Tags = PrepareDefaultKricoTags(memcachedExecutorConfig)
	kricoSnapTaskConfig.Tags["category"] = workload.TypeCaching
	kricoSnapTaskConfig.Tags["memory"] = strconv.FormatFloat(memory, 'f', -1, 64)
	kricoSnapTaskConfig.Tags["ratio"] = strconv.FormatFloat(ratio, 'f', -1, 64)
	kricoSnapTaskConfig.Tags["clients"] = strconv.FormatFloat(clients, 'f', -1, 64)

	// Prepare launcher
	kricoSnapTaskLauncher, err := kricosnapsession.NewSessionLauncher(kricoSnapTaskConfig)
	errutil.CheckWithContext(err, "Cannot obtain Snap Task Launcher")

	// Run Snap task
	kricoSnapTaskHandle, err := kricoSnapTaskLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics!")

	// Stop on the end
	defer kricoSnapTaskHandle.Stop()

	//
	// Aggressor
	//

	// Prepare executor
	mutilateExecutorConfig := executor.DefaultRemoteConfig()
	mutilateExecutor, err := executor.NewRemote(aggressorAddress.Value(), mutilateExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare Mutilate executor!")

	// Prepare launcher
	mutilateLauncher := mutilate.New(mutilateExecutor, mutilateConfig)

	// Start stressing workload
	mutilateTask, err := mutilateLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start Mutilate task!")

	// Stop stressing workload on the end
	defer mutilateTask.Stop()

	// Wait until load generating finishes
	mutilateTask.Wait(0)

}