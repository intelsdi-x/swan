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

/*
Openstack executor under the hood is using these are two components:
- executor:
	- talks to Openstack cluster to create new instance,
	- starts watcher (by calling Execute),
	- runs in main goroutine,
	- gives back control to user by returning newly created taskHandle
- watcher:
	- responsible for monitoring state of Instance, passing to information to taskHandle
	- also in case of failure or part of cleaning up or when asked directry by taskHandle - delete instance
*/

package executor

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/extendedserverattributes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	executorName        = "Openstack executor"
	executorLogPrefix   = executorName + ":"
	taskHandleName      = "Openstack task handle"
	taskHandleLogPrefix = taskHandleName + ":"
	directoryPrefix     = "openstack"
)

var (
	flavorDiskFlag      = conf.NewIntFlag("flavor_disk", "Openstack flavor disk size [GB]", 10)
	flavorRAMFlag       = conf.NewIntFlag("flavor_ram", "Openstack flavor RAM size [MB]", 1024)
	flavorVCPUsFlag     = conf.NewIntFlag("flavor_vcpus", "Openstack flavor VCPUs", 1)
	imageFlag           = conf.NewStringFlag("image", "Name of image.", "cirros")
	userFlag            = conf.NewStringFlag("username", "Username", "cirros")
	sshKeyPathFlag      = conf.NewStringFlag("ssh_key", "SSH key path", "~/.ssh/id_rsa")
	keypairName         = conf.NewStringFlag("os_keypair_name", "Openstack Keypair Name", "swan")
	bootUpTimeOut       = conf.NewDurationFlag("vm_boot_up_timeout", "Virtual Machine boot up timeout", time.Second*30)
	hostAggregateIDFlag = conf.NewIntFlag("host_aggregate_id", "ID of host aggregate which VM must be running in", -1)
)

// DefaultOpenstackConfig creates default OpenStack config.
func DefaultOpenstackConfig(auth gophercloud.AuthOptions) OpenstackConfig {
	return OpenstackConfig{
		Auth: auth,
		Flavor: OpenstackFlavor{
			Disk:  flavorDiskFlag.Value(),
			RAM:   flavorRAMFlag.Value(),
			VCPUs: flavorVCPUsFlag.Value(),
		},
		Image:      imageFlag.Value(),
		User:       userFlag.Value(),
		SSHKeyPath: sshKeyPathFlag.Value(),
	}
}

// OpenstackFlavor defines OpenStack flavor data.
type OpenstackFlavor struct {
	Name  string
	Disk  int
	RAM   int
	VCPUs int
}

// OpenstackAuthConfig defines OpenStack authentication configuration data.
type OpenstackAuthConfig struct {
	Username   string
	Password   string
	TenantID   string
	DomainName string
	Endpoint   string
}

type hostAggregate struct {
	Name             string
	ConfigurationID  string
	AvailabilityZone string
	Disk             disk
	Ram              ram
	Cpu              cpu
}

type hypervisor struct {
	InstanceName string
	Address      string
}

type disk struct {
	Iops string
	Size string
}

type ram struct {
	Bandwidth string
	Size      string
}

type cpu struct {
	Performance string
	Threads     string
}

// OpenstackConfig defines OpenStack instance configuration data.
type OpenstackConfig struct {
	Auth          gophercloud.AuthOptions
	Flavor        OpenstackFlavor
	Image         string
	User          string
	SSHKeyPath    string
	Name          string
	ID            string
	Hypervisor    hypervisor
	HostAggregate hostAggregate
}

// Openstack defines OpenStack server configuration and client.
type Openstack struct {
	config *OpenstackConfig
	client *gophercloud.ServiceClient
}

// NewOpenstack creates OpenStack executor.
func NewOpenstack(config *OpenstackConfig) Executor {
	return Openstack{config: config}
}

// String returns user-friendly name of executor.
func (stack Openstack) String() string {
	return fmt.Sprintf("%s at %s", executorName, stack.config.Auth.IdentityEndpoint)
}

// Execute runs provided command on OpenStack cluster.
func (stack Openstack) Execute(command string) (TaskHandle, error) {
	provider, err := openstack.AuthenticatedClient(stack.config.Auth)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get provider", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	stack.client, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get compute client", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	flavorID, err := stack.ensureFlavor()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err = stack.ensureKeypair(); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	floatingIP, err := stack.findFloatingIP()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	image, err := images.IDFromName(stack.client, stack.config.Image)

	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get image id from name: %s !", executorLogPrefix, image)
		log.Error(err.Error())
		return nil, err
	}

	instanceName := fmt.Sprintf("krico.%s", uuid.New())

	stack.config.Name = instanceName

	aggregate, err := aggregates.Get(stack.client, hostAggregateIDFlag.Value()).Extract()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get host aggregate info", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	stack.config.HostAggregate.Name = aggregate.Name
	stack.config.HostAggregate.AvailabilityZone = aggregate.AvailabilityZone

	stack.config.HostAggregate.ConfigurationID = aggregate.Metadata["configuration_id"]

	stack.config.HostAggregate.Disk.Iops = aggregate.Metadata["disk_iops"]
	stack.config.HostAggregate.Disk.Size = aggregate.Metadata["disk_size"]

	stack.config.HostAggregate.Cpu.Performance = aggregate.Metadata["cpu_performance"]
	stack.config.HostAggregate.Cpu.Threads = aggregate.Metadata["cpu_threads"]

	stack.config.HostAggregate.Ram.Size = aggregate.Metadata["ram_size"]
	stack.config.HostAggregate.Ram.Bandwidth = aggregate.Metadata["ram_bandwidth"]

	serverOpts := servers.CreateOpts{
		Name:             instanceName,
		FlavorRef:        flavorID,
		ImageRef:         image,
		AvailabilityZone: stack.config.HostAggregate.AvailabilityZone,
	}

	serverOptsExt := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           keypairName.Value(),
	}

	instance, err := servers.Create(stack.client, serverOptsExt).Extract()
	if err != nil {
		err = errors.Wrapf(err, "%s Unable to create instance", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Scheduled instance %s creation", executorLogPrefix, instance.ID)

	stack.config.ID = instance.ID

	if err = servers.WaitForStatus(stack.client, instance.ID, "ACTIVE", 60); err != nil {
		err = errors.Wrapf(err, "%s Couldn't launch instance", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Launched instance %s", executorLogPrefix, instance.ID)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: floatingIP,
	}

	if err = floatingips.AssociateInstance(stack.client, instance.ID, associateOpts).ExtractErr(); err != nil {
		err = errors.Wrapf(err, "%s Couldn't associate %q floating ip to instance %s", executorLogPrefix, floatingIP, instanceName)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Associated %q floating ip", executorLogPrefix, floatingIP)

	remoteConfig := RemoteConfig{
		User:    stack.config.User,
		KeyPath: stack.config.SSHKeyPath,
		Port:    22,
	}

	// Wait while to ensure that everything booted up
	log.Infof("%s Waiting for %s instance to boot up", executorLogPrefix, instanceName)
	time.Sleep(bootUpTimeOut.Value())

	stack.config.Hypervisor.InstanceName, err = stack.obtainHypervisorInstanceName(instance.ID)
	if err != nil {
		log.Errorf("%s Couldn't obtain hypervisor instance name for %s instance!", executorLogPrefix, instanceName)
		return nil, err
	}

	stack.config.Hypervisor.Address, err = stack.obtainHypervisorAddress(instance.ID)
	if err != nil {
		log.Errorf("%s Couldn't obtain hypervisor address for %s instance!", executorLogPrefix, instanceName)
		return nil, err
	}

	remote, err := NewRemote(floatingIP, remoteConfig)
	if err != nil {
		log.Errorf("%s Couldn't create remote executor for %s instance with %s ip", executorLogPrefix, instanceName, floatingIP)
		return nil, err
	}

	log.Infof("%s Created remote executor", executorLogPrefix)

	remoteHandler, err := remote.Execute(command)
	if err != nil {
		log.Errorf("%s Couldn't execute %q on %s (%s)", executorLogPrefix, command, instanceName, floatingIP)
		return nil, err
	}

	exitCode, _ := remoteHandler.ExitCode()

	log.Infof("%s Executed %q on %s (%s) with exit code: %d", executorLogPrefix, command, instanceName, floatingIP, exitCode)

	outputDirectory, err := createOutputDirectory(command, directoryPrefix)
	if err != nil {
		log.Errorf("%s Couldn't create output directory %q: %s", executorLogPrefix, directoryPrefix, err)
		return nil, err
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(outputDirectory)
	if err != nil {
		removeDirectory(outputDirectory)
		log.Errorf("%s Cannot create output files for command %q : %s", executorLogPrefix, command, err.Error())
		return nil, err
	}

	stdoutFileName := stdoutFile.Name()
	stderrFileName := stderrFile.Name()
	stdoutFile.Close()
	stderrFile.Close()

	taskHandle := &OpenstackTaskHandle{
		command:        command,
		hostIP:         floatingIP,
		stdoutFilePath: stdoutFileName,
		stderrFilePath: stderrFileName,
		instance:       instance.ID,
		os:             &stack,
		running:        true,
		exitCode:       exitCode,
		requestStop:    make(chan struct{}),
		requestDelete:  make(chan struct{}),
		stopped:        make(chan struct{}),
		deleted:        make(chan struct{}),
	}

	taskWatcher := &openstackWatcher{
		instance:      instance.ID,
		client:        stack.client,
		running:       &taskHandle.running,
		requestStop:   taskHandle.requestStop,
		requestDelete: taskHandle.requestDelete,
		stopped:       taskHandle.stopped,
		deleted:       taskHandle.deleted,
	}

	err = taskWatcher.watch()

	return taskHandle, nil
}

func (stack Openstack) ensureFlavor() (string, error) {
	stack.config.Flavor.Name = fmt.Sprintf("krico.cpu-%d.ram-%d.disk-%d", stack.config.Flavor.VCPUs, stack.config.Flavor.RAM, stack.config.Flavor.Disk)

	listOpts := flavors.ListOpts{
		AccessType: flavors.AllAccess,
	}

	allPages, err := flavors.ListDetail(stack.client, listOpts).AllPages()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get flavors list to check %s within it", executorLogPrefix, stack.config.Flavor.Name)
		return "", err
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't extract flavors list to check %s within it", executorLogPrefix, stack.config.Flavor.Name)
		return "", err
	}

	// check if flavor already exists
	for _, flavor := range allFlavors {
		if flavor.Name == stack.config.Flavor.Name {
			log.Infof("%s Using existing flavor %q", executorLogPrefix, stack.config.Flavor.Name)
			return flavor.ID, nil
		}
	}

	// create flavor
	flavorID := uuid.New()
	flavorOpts := flavors.CreateOpts{
		ID:    flavorID,
		Name:  stack.config.Flavor.Name,
		Disk:  &stack.config.Flavor.Disk,
		RAM:   stack.config.Flavor.RAM,
		VCPUs: stack.config.Flavor.VCPUs,
	}
	_, err = flavors.Create(stack.client, flavorOpts).Extract()
	if err != nil {
		err = errors.Wrapf(err, "Unable to create flavor %q", stack.config.Flavor.Name)
		return flavorID, err
	}

	log.Infof("%s Created flavor %q", executorLogPrefix, stack.config.Flavor.Name)

	return flavorID, nil
}

func (stack Openstack) ensureKeypair() error {
	allPages, err := keypairs.List(stack.client).AllPages()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't read keypairs list", executorLogPrefix)
		return err
	}

	allKeyPairs, err := keypairs.ExtractKeyPairs(allPages)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't extract keypairs list", executorLogPrefix)
		return err
	}

	// check if keypair already exists
	for _, kp := range allKeyPairs {
		if kp.Name == keypairName.Value() {
			log.Infof("%s Using existing keypair %q", executorLogPrefix, keypairName.Value())
			return nil
		}
	}

	// create keypair
	publicKeyPath := stack.config.SSHKeyPath + ".pub"
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't read public key %q", executorLogPrefix, publicKeyPath)
		return err
	}

	keypairOpts := keypairs.CreateOpts{
		Name:      keypairName.Value(),
		PublicKey: string(publicKey),
	}

	_, err = keypairs.Create(stack.client, keypairOpts).Extract()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't create keypair %q", executorLogPrefix, keypairName.Value())
		return err
	}

	log.Infof("%s Created keypair %q", executorLogPrefix, keypairName.Value())

	return nil
}

func (stack Openstack) findFloatingIP() (string, error) {
	allPages, err := floatingips.List(stack.client).AllPages()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't read floating IPs list", executorLogPrefix)
		return "", err
	}

	allFloatingIPs, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't extract floating IPs list", executorLogPrefix)
		return "", err
	}

	var floatingIP *floatingips.FloatingIP

	for _, fip := range allFloatingIPs {
		if fip.FixedIP == "" {
			floatingIP = &fip
			break
		}
	}

	if floatingIP == nil {
		err = errors.Errorf("%s Couldn't find free floating ip", executorLogPrefix)
		return "", err
	}

	return floatingIP.IP, nil
}

type serverAttributesExt struct {
	servers.Server
	extendedserverattributes.ServerAttributesExt
}

func (stack Openstack) obtainHypervisorInstanceName(instanceID string) (string, error) {

	var extendedAttributes serverAttributesExt

	err := servers.Get(stack.client, instanceID).ExtractInto(&extendedAttributes)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get %s instance extended attributes", executorLogPrefix, instanceID)
		return "", err
	}

	return extendedAttributes.InstanceName, nil
}

func (stack Openstack) obtainHypervisorAddress(instanceID string) (string, error) {

	var extendedAttributes serverAttributesExt

	err := servers.Get(stack.client, instanceID).ExtractInto(&extendedAttributes)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get %s instance extended attributes", executorLogPrefix, instanceID)
		return "", err
	}

	allPages, err := hypervisors.List(stack.client).AllPages()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't read hypervisors list", executorLogPrefix)
		return "", err
	}

	allHypervisors, err := hypervisors.ExtractHypervisors(allPages)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't extract hypervisors list", executorLogPrefix)
		return "", err
	}

	var hypervisorAddress string

	for _, hypervisor := range allHypervisors {
		if hypervisor.HypervisorHostname == extendedAttributes.HypervisorHostname {
			hypervisorAddress = hypervisor.HostIP
			break
		}
	}

	if hypervisorAddress == "" {
		err = errors.Errorf("%s Couldn't find hypervisor address!", executorLogPrefix)
		return "", err
	}

	return hypervisorAddress, err
}

// OpenstackTaskHandle represents an abstraction to control task lifecycle and status.
type OpenstackTaskHandle struct {
	command        string
	hostIP         string
	stdoutFilePath string
	stderrFilePath string
	instance       string
	os             *Openstack
	running        bool
	exitCode       int
	requestStop    chan struct{}
	requestDelete  chan struct{}
	stopped        chan struct{}
	deleted        chan struct{}
}

// String returns user-friendly name of task handle.
func (th *OpenstackTaskHandle) String() string {
	return fmt.Sprintf("Openstack instance %q running with command %q on %s", th.instance, th.command, th.hostIP)
}

// Address returns ip address of host where task is located.
func (th *OpenstackTaskHandle) Address() string {
	return th.hostIP
}

// EraseOutput removes directory where output files resides.
func (th *OpenstackTaskHandle) EraseOutput() error {
	outputDir := filepath.Dir(th.stdoutFilePath)
	return removeDirectory(outputDir)
}

// ExitCode returns exit code of finished task.
// If task is still running, return error.
func (th *OpenstackTaskHandle) ExitCode() (int, error) {

	if th.isRunning() {
		err := errors.Errorf("Task is still running")
		log.Error(err.Error())
		return 0, err
	}

	return th.exitCode, nil
}

// Status returns task status.
func (th *OpenstackTaskHandle) Status() TaskState {
	if th.isRunning() {
		return RUNNING
	}

	return TERMINATED
}

// StderrFile returns file handle for file to the task's stderr file.
func (th *OpenstackTaskHandle) StderrFile() (*os.File, error) {
	return openFile(th.stderrFilePath)
}

// StdoutFile returns file handle for file to the task's stdout file.
func (th *OpenstackTaskHandle) StdoutFile() (*os.File, error) {
	return openFile(th.stdoutFilePath)
}

// Stop stops task.
func (th *OpenstackTaskHandle) Stop() error {
	if !th.isRunning() {
		return nil
	}

	log.Debugf("%s delete instance %q", taskHandleLogPrefix, th.instance)

	th.requestStop <- struct{}{}

	log.Debugf("%s waiting for instance %q to stop", taskHandleLogPrefix, th.instance)

	<-th.stopped

	log.Debugf("%s instance %q stop", taskHandleLogPrefix, th.instance)

	th.requestDelete <- struct{}{}

	<-th.deleted

	log.Debugf("%s instance %q deleted", taskHandleLogPrefix, th.instance)

	return nil
}

// Wait blocks and waits for task to terminate.
// For '0' it'll wait until task termination.
func (th *OpenstackTaskHandle) Wait(timeout time.Duration) (bool, error) {
	if !th.isRunning() {
		return true, nil
	}

	timeoutChannel := getTimeoutChan(timeout)

	select {
	case <-th.stopped:
		return true, nil
	case <-timeoutChannel:
		return false, nil
	}
}

func (th *OpenstackTaskHandle) isRunning() bool {
	return th.running
}

// Instance returns OpenStack instance name.
func (th *OpenstackTaskHandle) Instance() string {
	return th.instance
}

type openstackWatcher struct {
	client        *gophercloud.ServiceClient
	instance      string
	running       *bool
	requestStop   chan struct{}
	requestDelete chan struct{}
	stopped       chan struct{}
	deleted       chan struct{}
}

func (watcher *openstackWatcher) watch() error {

	go func() {
		for {
			server, _ := servers.Get(watcher.client, watcher.instance).Extract()
			switch server.Status {
			case "ACTIVE":
				*watcher.running = true
			case "BUILDING":
				*watcher.running = true
			case "DELETED":
				*watcher.running = false
			case "ERROR":
				*watcher.running = false
				watcher.requestDelete <- struct{}{}
			case "SHUTOFF":
				*watcher.running = false
				watcher.stopped <- struct{}{}
			case "PAUSED":
				*watcher.running = false
				*watcher.running = false
			case "RESCUED":
				*watcher.running = true
			case "STOPPED":
				*watcher.running = false
			case "SOFT_DELETED":
				*watcher.running = false
				watcher.requestDelete <- struct{}{}
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			select {
			case <-watcher.requestStop:
				startstop.Stop(watcher.client, watcher.instance).ExtractErr()
			case <-watcher.requestDelete:
				servers.Delete(watcher.client, watcher.instance)
				watcher.deleted <- struct{}{}
			}
		}
	}()

	return nil
}
