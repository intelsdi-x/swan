package workload

import (
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	kricosnapsession "github.com/intelsdi-x/swan/pkg/snap/sessions/krico"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/intelsdi-x/swan/pkg/workloads/redis"
	"github.com/intelsdi-x/swan/pkg/workloads/ycsb"
	"strconv"
)

// CollectingMetricsForCachingWorkload runs metric gathering experiment for caching workload.
func CollectingMetricsForCachingWorkload(experimentID string) {

	//	Load OpenStack authentication variables from environment.
	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	//
	//	Workload
	//

	//	Prepare configuration.
	workloadConfig := memcached.DefaultMemcachedConfig()

	//	Prepare executor.
	workloadExecutorConfig := executor.DefaultOpenstackConfig(auth)
	workloadExecutorConfig.Image = "krico_memcached"
	workloadExecutor := executor.NewOpenstack(&workloadExecutorConfig)

	//	Prepare launcher.
	workloadLauncher := memcached.New(workloadExecutor, workloadConfig)

	//	Run workload.
	workloadHandle, err := workloadLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot launch Memcached!")

	//	Stop workload in the end of experiment.
	defer workloadHandle.Stop()

	//
	//	Load generator
	//

	//	Prepare configuration.
	loadGeneratorConfig := mutilate.DefaultMutilateConfig()
	loadGeneratorConfig.MemcachedHost = workloadHandle.Address()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	//	Prepare executor.
	loadGeneratorExecutorConfig := executor.DefaultRemoteConfig()
	loadGeneratorExecutor, err := executor.NewRemote(aggressorAddress.Value(), loadGeneratorExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare Mutilate executor!")

	//	Prepare launcher.
	loadGeneratorLauncher := mutilate.New(loadGeneratorExecutor, loadGeneratorConfig)

	//	Populate workload.
	err = loadGeneratorLauncher.Populate()
	errutil.CheckWithContext(err, "Cannot load the initial test data into Memcached!")

	//
	//	Metrics collector (Snap)
	//

	//	Start Snap on OpenStack host.
	StartSnapService(workloadExecutorConfig.Hypervisor.Address)

	//	Obtain workload instance cgroup.
	cgroup, err := GetInstanceCgroup(workloadExecutorConfig.Hypervisor.InstanceName, workloadExecutorConfig.Hypervisor.Address)
	errutil.CheckWithContext(err, "Cannot obtain instance cgroup!")

	//
	//	KRICO parameters
	//

	//	Calculate workload parameters.
	memory := float64(workloadConfig.MaxMemoryMB / 1024) // total cache size [GiB]
	ratio := loadGeneratorConfig.Update                  // estimated get vs set ratio [0.0 - 1.0]
	clients := float64(loadGeneratorConfig.MasterThreads * loadGeneratorConfig.MasterConnections)

	//
	//	Snap task
	//

	//	Configure task.
	snapTaskConfig := kricosnapsession.DefaultConfig(cgroup, workloadExecutorConfig.Hypervisor.InstanceName)
	snapTaskConfig.Tags = PrepareDefaultKricoTags(workloadExecutorConfig, experimentID)
	snapTaskConfig.Tags["category"] = TypeCaching
	snapTaskConfig.Tags["memory"] = strconv.FormatFloat(memory, 'f', -1, 64)
	snapTaskConfig.Tags["ratio"] = strconv.FormatFloat(ratio, 'f', -1, 64)
	snapTaskConfig.Tags["clients"] = strconv.FormatFloat(clients, 'f', -1, 64)

	//	Prepare launcher.
	snapTaskLauncher, err := kricosnapsession.NewSessionLauncher(snapTaskConfig)
	errutil.CheckWithContext(err, "Cannot obtain Snap Task Launcher !")

	//	Run Snap task.
	snapTaskHandle, err := snapTaskLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics!")

	//	Stop on the end.
	defer snapTaskHandle.Stop()

	//
	//	Load generator
	//

	//	Start load on workload.
	loadGeneratorHandle, err := loadGeneratorLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start Mutilate task!")

	//	In the end stop load on workload.
	defer loadGeneratorHandle.Stop()

	//	Wait until load generating finishes.
	loadGeneratorHandle.Wait(0)
}

// ClassifyCachingWorkload runs classify experiment for caching workload.
func ClassifyCachingWorkload(experimentID string) string {

	//	Load OpenStack authentication variables from environment.
	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	//
	//	Workload
	//

	//	Prepare configuration.
	workloadConfig := redis.DefaultConfig()

	//	Prepare executor.
	workloadExecutorConfig := executor.DefaultOpenstackConfig(auth)
	workloadExecutorConfig.Image = "krico_redis"
	workloadExecutor := executor.NewOpenstack(&workloadExecutorConfig)

	//	Prepare launcher.
	workloadLauncher := redis.New(workloadExecutor, workloadConfig)

	//	Run workload.
	workloadHandle, err := workloadLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot launch Redis!")

	//	Stop workload in the end of experiment.
	defer workloadHandle.Stop()

	//
	//	Load generator
	//

	//	Prepare configuration.
	loadGeneratorConfig := ycsb.DefaultYcsbConfig()
	loadGeneratorConfig.RedisHost = workloadHandle.Address()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	//	Prepare executor.
	loadGeneratorExecutorConfig := executor.DefaultRemoteConfig()
	loadGeneratorExecutor, err := executor.NewRemote(aggressorAddress.Value(), loadGeneratorExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare YCSB Redis executor !")

	//	Prepare launcher.
	ycsb.CalculateWorkloadCommandParameters(maxQPS, loadDuration, &loadGeneratorConfig)
	loadGeneratorLauncher := ycsb.New(loadGeneratorExecutor, loadGeneratorConfig)

	//	Populate workload.
	err = loadGeneratorLauncher.Populate()
	errutil.CheckWithContext(err, "Cannot load the initial test data into Redis!")

	//
	//	Metrics collector (Snap)
	//

	//	Start Snap on OpenStack host.
	StartSnapService(workloadExecutorConfig.Hypervisor.Address)

	//	Obtain workload instance cgroup.
	cgroup, err := GetInstanceCgroup(workloadExecutorConfig.Hypervisor.InstanceName, workloadExecutorConfig.Hypervisor.Address)
	errutil.CheckWithContext(err, "Cannot obtain workload instance cgroup !")

	snapTaskConfig := kricosnapsession.DefaultConfig(cgroup, workloadExecutorConfig.Hypervisor.InstanceName)
	snapTaskConfig.Tags = PrepareDefaultKricoTags(workloadExecutorConfig, experimentID)

	//	Prepare launcher.
	snapTaskLauncher, err := kricosnapsession.NewSessionLauncher(snapTaskConfig)
	errutil.CheckWithContext(err, "Cannot obtain Snap Task Launcher !")

	//	Run Snap task.
	snapTaskHandle, err := snapTaskLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics !")

	//	Stop snap task on the end of experiment.
	defer snapTaskHandle.Stop()

	//
	//	Load Generator
	//

	//	Start load on workload.
	loadGeneratorHandle, err := loadGeneratorLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start YCSB Redis task !")

	//	In the end stop load on workload.
	defer loadGeneratorHandle.Stop()

	//	Wait until load generating finishes.
	loadGeneratorHandle.Wait(0)

	return workloadExecutorConfig.ID
}
