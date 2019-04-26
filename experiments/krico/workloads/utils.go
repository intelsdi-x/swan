package workload

import (
	"fmt"
	"strings"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/libvirt/libvirt-go"
	"github.com/pkg/errors"
)

const (
	// TypeCaching is constant name of caching workload.
	TypeCaching = "caching"
)

var (
	snapStartCommand = "systemctl restart snap-telemetry"
	aggressorAddress = conf.NewStringFlag("aggressor_address", "IP address of aggressor node.", "127.0.0.0")
)

// RunCollectingMetrics runs metric gathering experiment for each type of workload.
func RunCollectingMetrics(experimentID string) {
	CollectingMetricsForCachingWorkload(experimentID)
}

// RunWorkloadsClassification runs classification experiment for each type of workload. Return instances id.
func RunWorkloadsClassification(experimentID string) []string {
	var instances []string

	instances = append(instances, ClassifyCachingWorkload(experimentID))

	return instances
}

// StartSnapService starts Snap Telemetry Framework.
func StartSnapService(address string) error {

	startSnapExecutor, err := executor.NewRemoteFromIP(address)
	if err != nil {
		err = errors.Wrapf(err, "Cannot obtain Snap-telemetry executor on %q !")
		return err
	}

	startSnapTaskHandle, err := startSnapExecutor.Execute(snapStartCommand)
	if err != nil {
		err = errors.Wrapf(err, "Cannot execute %q command on %q !", snapStartCommand, address)
		return err
	}

	//	Wait until start
	_, err = startSnapTaskHandle.Wait(0)
	if err != nil {
		err = errors.Wrap(err, "Cannot start Snap-telemetry service !")
		return err
	}

	return nil
}

// GetInstanceCgroup provides cgroup of libvirt instance.
func GetInstanceCgroup(hypervisorInstanceName string, hypervisorAddress string) (string, error) {

	conn, err := libvirt.NewConnectReadOnly("qemu+ssh://root@" + hypervisorAddress + "/system")
	if err != nil {
		return "", errors.Wrap(err, "Couldn't connect to Libvirt!")
	}
	defer conn.Close()

	domain, err := conn.LookupDomainByName(hypervisorInstanceName)
	if err != nil {
		return "", errors.Wrap(err, "Couldn't get instance domain!")
	}

	instanceID, err := domain.GetID()
	if err != nil {
		return "", errors.Wrap(err, "Couldn't get instance domain id!")
	}

	instanceName := strings.Replace(hypervisorInstanceName, "-", `\x2d`, -1)

	cgroup := "machine.slice:machine-qemu" + `\x2d` + fmt.Sprint(instanceID) + `\x2d` + instanceName + ".scope"

	return cgroup, nil
}

// PrepareDefaultKricoTags returns struct with default tags needed in KRICO experiment.
func PrepareDefaultKricoTags(openStackConfig executor.OpenstackConfig, experimentID string) map[string]interface{} {
	return map[string]interface{}{
		experiment.ExperimentKey:          experimentID,
		"name":                            openStackConfig.Name,
		"instance_id":                     openStackConfig.ID,
		"image":                           openStackConfig.Image,
		"flavor_name":                     openStackConfig.Flavor.Name,
		"flavor_disk":                     openStackConfig.Flavor.Disk,
		"flavor_ram":                      openStackConfig.Flavor.RAM,
		"flavor_vcpus":                    openStackConfig.Flavor.VCPUs,
		"host_aggregate_name":             openStackConfig.HostAggregate.Name,
		"host_aggregate_configuration_id": openStackConfig.HostAggregate.ConfigurationID,
		"host_aggregate_disk_iops":        openStackConfig.HostAggregate.Disk.Iops,
		"host_aggregate_disk_size":        openStackConfig.HostAggregate.Disk.Size,
		"host_aggregate_ram_bandwidth":    openStackConfig.HostAggregate.RAM.Bandwidth,
		"host_aggregate_ram_size":         openStackConfig.HostAggregate.RAM.Size,
		"host_aggregate_cpu_performance":  openStackConfig.HostAggregate.CPU.Performance,
		"host_aggregate_cpu_threads":      openStackConfig.HostAggregate.CPU.Threads,
	}
}
