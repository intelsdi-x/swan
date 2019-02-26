package workload

import (
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/intelsdi-x/backup/swan/experiments/krico/workloads"
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

func CollectingMetricsForCachingWorkload() {

	//	Load OpenStack authentication variables from environment
	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	//
	//	Workload
	//

	//	Prepare configuration
	workloadConfig := memcached.DefaultMemcachedConfig()

	//	Prepare executor
	workloadExecutorConfig := executor.DefaultOpenstackConfig(auth)
	workloadExecutorConfig.Image = "krico_memcached"
	workloadExecutor := executor.NewOpenstack(&workloadExecutorConfig)

	//	Prepare launcher
	workloadLauncher := memcached.New(workloadExecutor, workloadConfig)

	//	Run workload
	workloadHandle, err := workloadLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot launch Memcached!")

	//	Stop workload in the end of experiment
	defer workloadHandle.Stop()

	//
	//	Aggressor
	//

	//	Prepare configuration
	aggressorConfig := mutilate.DefaultMutilateConfig()
	aggressorConfig.MemcachedHost = workloadHandle.Address()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	//	Prepare executor
	aggressorExecutorConfig := executor.DefaultRemoteConfig()
	aggressorExecutor, err := executor.NewRemote(aggressorAddress.Value(), aggressorExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare Mutilate executor!")

	//	Prepare launcher
	aggressorLauncher := mutilate.New(aggressorExecutor, aggressorConfig)

	//	Populate workload.
	err = aggressorLauncher.Populate()
	errutil.CheckWithContext(err, "Cannot load the initial test data into Memcached!")

	//
	//	Metrics collector (Snap)
	//

	//	Start Snap on OpenStack host.
	StartSnapService(workloadExecutorConfig.Hypervisor.Address)

	//	Obtain workload instance cgroup
	cgroup, err := GetInstanceCgroup(workloadExecutorConfig.Hypervisor.InstanceName, workloadExecutorConfig.Hypervisor.Address)
	errutil.CheckWithContext(err, "Cannot obtain instance cgroup!")

	//
	//	KRICO parameters
	//

	//	Calculate workload parameters
	memory := float64(workloadConfig.MaxMemoryMB / 1024)         // total cache size [GiB]
	ratio, err := strconv.ParseFloat(aggressorConfig.Update, 64) // estimated get vs set ratio [0.0 - 1.0]
	errutil.CheckWithContext(err, "Cannot parse mutilate update argument (ratio parameter)!")
	clients := float64(aggressorConfig.MasterThreads * aggressorConfig.MasterConnections)

	//
	//	Snap task
	//

	//	Configure task
	snapTaskConfig := kricosnapsession.DefaultConfig(cgroup, workloadExecutorConfig.Hypervisor.InstanceName)
	snapTaskConfig.Tags = PrepareDefaultKricoTags(workloadExecutorConfig)
	snapTaskConfig.Tags["category"] = workload.TypeCaching
	snapTaskConfig.Tags["memory"] = strconv.FormatFloat(memory, 'f', -1, 64)
	snapTaskConfig.Tags["ratio"] = strconv.FormatFloat(ratio, 'f', -1, 64)
	snapTaskConfig.Tags["clients"] = strconv.FormatFloat(clients, 'f', -1, 64)

	//	Prepare launcher
	snapTaskLauncher, err := kricosnapsession.NewSessionLauncher(snapTaskConfig)
	errutil.CheckWithContext(err, "Cannot obtain Snap Task Launcher !")

	//	Run Snap task
	snapTaskHandle, err := snapTaskLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics!")

	//	Stop on the end
	defer snapTaskHandle.Stop()

	//
	//	Aggressor
	//

	//	Start stressing workload
	aggressorHandle, err := aggressorLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start Mutilate task!")

	//	Stop stressing workload on the end
	defer aggressorHandle.Stop()

	//	Wait until load generating finishes
	aggressorHandle.Wait(0)

}

func ClassifyCachingWorkload() {

	//	Load OpenStack authentication variables from environment
	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	//
	//	Workload
	//

	//	Prepare configuration
	workloadConfig := redis.DefaultRedisConfig()

	//	Prepare executor
	workloadExecutorConfig := executor.DefaultOpenstackConfig(auth)
	workloadExecutorConfig.Image = "krico_redis"
	workloadExecutor := executor.NewOpenstack(&workloadExecutorConfig)

	//	Prepare launcher
	workloadLauncher := redis.New(workloadExecutor, workloadConfig)

	//	Run workload
	workloadHandle, err := workloadLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot launch Redis!")

	//	Stop workload in the end of experiment
	//defer workloadHandle.Stop()

	//
	//	Aggressor
	//

	//	Prepare configuration
	aggressorConfig := ycsb.DefaultYcsbConfig()
	aggressorConfig.RedisHost = workloadHandle.Address()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	//
	//	Metrics collector (Snap)
	//

	//	Start Snap on OpenStack host.
	StartSnapService(workloadExecutorConfig.Hypervisor.Address)

	//	Obtain workload instance cgroup
	cgroup, err := GetInstanceCgroup(workloadExecutorConfig.Hypervisor.InstanceName, workloadExecutorConfig.Hypervisor.Address)
	errutil.CheckWithContext(err, "Cannot obtain workload instance cgroup !")

	snapTaskConfig := kricosnapsession.DefaultConfig(cgroup, workloadExecutorConfig.Hypervisor.InstanceName)
	snapTaskConfig.Tags = PrepareDefaultKricoTags(workloadExecutorConfig)

	//	Prepare launcher
	snapTaskLauncher, err := kricosnapsession.NewSessionLauncher(snapTaskConfig)
	errutil.CheckWithContext(err, "Cannot obtain Snap Task Launcher !")

	//	Run Snap task
	snapTaskHandle, err := snapTaskLauncher.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics !")

	//	Stop snap task on the end of experiment
	defer snapTaskHandle.Stop()

	//
	//	Aggressor
	//

	//	Prepare executor
	aggressorExecutorConfig := executor.DefaultRemoteConfig()
	aggressorExecutor, err := executor.NewRemote(aggressorAddress.Value(), aggressorExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare YCSB Redis executor !")

	//	Prepare launcher
	aggressorLauncher := ycsb.New(aggressorExecutor, aggressorConfig)

	//	Start stressing workload
	aggressorHandle, err := aggressorLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start YCSB Redis task !")

	//	Stop stressing workload on the end of experiment
	defer aggressorHandle.Stop()

	//	Wait until load generating finishes
	aggressorHandle.Wait(0)
}
