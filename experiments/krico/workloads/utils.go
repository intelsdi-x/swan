package workload

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/libvirt/libvirt-go"
	"github.com/pkg/errors"
	"strings"
)

const (
	TypeBigdata    = "bigdata"
	TypeCaching    = "caching"
	TypeOltp       = "oltp"
	TypeScience    = "science"
	TypeStreaming  = "streaming"
	TypeWebserving = "webserving"
)

var (
	experimentID     string
	snapStartCommand = "service snap-telemetry restart"
	aggressorAddress = conf.NewStringFlag("aggressor_address", "IP address of aggressor node.", "127.0.0.0")
)

func Initialize(experimentId string) {
	experimentID = experimentId
}

func RunCollectingMetrics() {
	CollectingMetricsForCachingWorkload()
}

func RunWorkloadsClassification() {
	ClassifyCachingWorkload()
}

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

func GetInstanceCgroup(hypervisorInstanceName string, hypervisorAddress string) (string, error) {

	conn, err := libvirt.NewConnectReadOnly("qemu+ssh://root@" + hypervisorAddress + "/system")
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

	instanceName := strings.Replace(hypervisorInstanceName, "-", `\x2d`, -1)

	cgroup := "machine.slice:machine-qemu" + `\x2d` + fmt.Sprint(instanceId) + `\x2d` + instanceName + ".scope"

	return cgroup, nil
}

func PrepareDefaultKricoTags(openStackConfig executor.OpenstackConfig) map[string]interface{} {
	return map[string]interface{}{
		experiment.ExperimentKey: experimentID,
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
		"host_aggregate_ram_bandwidth":    openStackConfig.HostAggregate.Ram.Bandwidth,
		"host_aggregate_ram_size":         openStackConfig.HostAggregate.Ram.Size,
		"host_aggregate_cpu_performance":  openStackConfig.HostAggregate.Cpu.Performance,
		"host_aggregate_cpu_threads":      openStackConfig.HostAggregate.Cpu.Threads,
	}
}
