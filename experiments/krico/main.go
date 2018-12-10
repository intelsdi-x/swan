// Copyright (c) 2018 Intel Corporation
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

package main

import (
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/metadata"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	kricosnapsession "github.com/intelsdi-x/swan/pkg/snap/sessions/krico"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/libvirt/libvirt-go"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/experiments/krico/workloads"
)

var (
	appName = os.Args[0]
	aggressorAddress = conf.NewStringFlag("aggressor_address", "IP address of aggresor node.", "127.0.0.0")
	snapAddress = conf.NewStringFlag("snap_address", "IP address of metric gathering node.", "127.0.0.0")
	hypervisorAddress = conf.NewStringFlag("hypervisor_address", "IP address of Openstack hypervisor node.", "127.0.0.0")

)

func main() {

	// Preparing application - setting name, help, parsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, uid)

	// Connect to metadata database.
	metaData, err := metadata.NewDefault(uid)
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(metaData, experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Read configuration.
	loadDuration := sensitivity.LoadDurationFlag.Value()
	maxQPS := sensitivity.PeakLoadFlag.Value()

	auth, err := openstack.AuthOptionsFromEnv()
	errutil.CheckWithContext(err, "Cannot read OpenStack environment variables!")

	// Prepare Memcached workload
	memcachedExecutorConfig := executor.DefaultOpenstackConfig(auth)
	memcachedExecutor := executor.NewOpenstack(&memcachedExecutorConfig)

	memcachedConfig := memcached.DefaultMemcachedConfig()
	memcachedLauncher := memcached.New(memcachedExecutor,memcachedConfig)

	// Run Memcached workload
	memcachedWorkload, err := memcachedExecutor.Execute(memcachedLauncher.BuildCommand()+" -d")
	errutil.CheckWithContext(err, "Cannot launch Memcached")

	// In the end, stop workload
	defer memcachedWorkload.Stop()

	// Get workload cgroup.
	cgroup, err := GetInstanceCgroup(memcachedExecutorConfig.HypervisorInstanceName, hypervisorAddress.Value())
	errutil.CheckWithContext(err, "Cannot obtain Memcached cgroup for Snap.")

	// Start Snap on Openstack host.
	// Cannot do this earlier because Snap prepare available metrics on start.
	// Otherwise it wouldn't see metrics from workload cgroup.
	setupSnapExecutor, err := executor.NewRemoteFromIP(snapAddress.Value())
	errutil.CheckWithContext(err, "Cannot obtain Snap executor.")

	// Restart because:
	//			- if service stopped, it would start
	//			- if service running, it would start again
	setupSnapTaskHandle, err:= setupSnapExecutor.Execute("service snap-telemetry restart")
	errutil.CheckWithContext(err, "Cannot execute start command on bare metal with Snap.")

	// Wait until Snap start.
	_, err = setupSnapTaskHandle.Wait(0)
	errutil.CheckWithContext(err, "Cannot start Snap telemetry service.")

	// Wait until Snap load plugins.
	time.Sleep(time.Second*10)

	// Prepare Snap task.
	kricoSnapConfig := kricosnapsession.DefaultConfig(cgroup, memcachedExecutorConfig.HypervisorInstanceName)
	kricoSnapConfig.Tags = map[string]interface{}{
		experiment.ExperimentKey: uid,
		"configuration_id": memcachedExecutorConfig.FlavorName,
	}

	kricoSnapSession, err := kricosnapsession.NewSessionLauncher(kricoSnapConfig)
	errutil.CheckWithContext(err, "KRICO telemetry collection failed")

	// Run Snap task.
	kricoSnapSessionHandle, err := kricoSnapSession.Launch()
	errutil.CheckWithContext(err, "Cannot gather performance metrics!")

	// In the experiment end, stop task.
	defer kricoSnapSessionHandle.Stop()

	// Prepare Mutilate.
	mutilateExecutorConfig := executor.DefaultRemoteConfig()

	mutilateExecutor, err := executor.NewRemote(aggressorAddress.Value(), mutilateExecutorConfig)
	errutil.CheckWithContext(err, "Cannot prepare Mutilate executor!")

	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedWorkload.Address()

	mutilateLauncher := mutilate.New(mutilateExecutor, mutilateConfig)

	// Start stressing Memcached with Mutilate.
	mutilateTask, err := mutilateLauncher.Load(maxQPS, loadDuration)
	errutil.CheckWithContext(err, "Cannot start Mutilate task!")

	// In the experiment end, stop stressing.
	defer mutilateTask.Stop()

	// Waiting until load generating finishes.
	mutilateTask.Wait(0)

	// Calculate parameters
	memory := float32(memcachedConfig.MaxMemoryMB / 1024) // <float> total cache size [GiB]
	// TODO: Change SWAN mutilate handle to set -u/-update
	ratio := 0.0 // <float> estimated get vs. set ratio [0.0 - 1.0]
	client := mutilateConfig.MasterThreads*mutilateConfig.MasterConnections // number of concurrent client connections

	// Prepare metadata.
	records := map[string]string{
		"commands_arguments": strings.Join(os.Args, ","),
		"experiment_name":    appName,
		"load_duration":      loadDuration.String(),
		"max_qps":            strconv.Itoa(maxQPS),
		"flavor_disk":		  strconv.Itoa(memcachedExecutorConfig.Flavor.Disk),
		"flavor_ram":         strconv.Itoa(memcachedExecutorConfig.Flavor.RAM),
		"flavor_vcpus":       strconv.Itoa(memcachedExecutorConfig.Flavor.VCPUs),
		"image":              memcachedExecutorConfig.Image,
		"category":		      workload.TypeCaching,
		"memory":			  fmt.Sprintf("%f", memory),
		"ratio":			  fmt.Sprintf("%f", ratio),
		"client":			  strconv.Itoa(client),
	}

	// Save metadata.
	err = metaData.RecordMap(records, metadata.TypeEmpty)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")
}

func GetInstanceCgroup(hypervisorInstanceName string, hypervisorAddress string) (string, error){
	conn, err := libvirt.NewConnect("qemu+ssh://root@"+hypervisorAddress+"/system")
	if err != nil {
		return "", fmt.Errorf("couldn't connect to libvirt: %v", err)
	}
	defer conn.Close()

	domain, err := conn.LookupDomainByName(hypervisorInstanceName)
	if err != nil {
		return "", fmt.Errorf("couldn't get instance domain: %v", err)
	}

	instanceId, err := domain.GetID()
	if err != nil {
		return "", fmt.Errorf("couldn't get instance domain id: %v", err)
	}

	instanceName := strings.Replace(hypervisorInstanceName, "-", `\x2d`,-1)

	cgroup := "machine.slice:machine-qemu"+`\x2d`+fmt.Sprint(instanceId)+`\x2d`+instanceName+".scope"

	return cgroup, nil
}
