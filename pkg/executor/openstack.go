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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/extendedserverattributes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	executorName        = "Openstack executor"
	executorLogPrefix   = executorName + ":"
	taskHandleName      = "Openstack task handle"
	taskHandleLogPrefix = taskHandleName + ":"
)

var (
	flavorDiskFlag  = conf.NewIntFlag("flavor_disk", "Openstack flavor disk size [GB]", 10)
	flavorRAMFlag   = conf.NewIntFlag("flavor_ram", "Openstack flavor RAM size [MB]", 1024)
	flavorVCPUsFlag = conf.NewIntFlag("flavor_vcpus", "Openstack flavor VCPUs", 1)
	imageFlag       = conf.NewStringFlag("image", "Name of image.", "cirros")
	userFlag        = conf.NewStringFlag("username", "Username", "cirros")
	sshKeyPathFlag  = conf.NewStringFlag("ssh_key", "SSH key path", "~/.ssh/id_rsa")
	keypairName     = conf.NewStringFlag("os_keypair_name", "Openstack Keypair Name", "swan")
	bootUpTimeOut   = conf.NewDurationFlag("vm_boot_up_timeout", "Virtual Machine boot up timeout", time.Second*30)
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

// OpenstackConfig defines OpenStack instance configuration data.
type OpenstackConfig struct {
	Auth                   gophercloud.AuthOptions
	Flavor                 OpenstackFlavor
	FlavorName             string
	Image                  string
	User                   string
	SSHKeyPath             string
	HypervisorInstanceName string
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
func (os Openstack) String() string {
	return fmt.Sprintf("%s at %s", executorName, os.config.Auth.IdentityEndpoint)
}

// Execute runs provided command on OpenStack cluster.
func (os Openstack) Execute(command string) (TaskHandle, error) {
	provider, err := openstack.AuthenticatedClient(os.config.Auth)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get provider", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	os.client, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get compute client", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	flavorID, err := os.ensureFlavor()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't ensure flavor", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	if err = os.ensureKeypair(); err != nil {
		err = errors.Wrapf(err, "%s Couldn't ensure keypair", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	floatingIP, err := os.findFloatingIP()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't find floating IP", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	image, err := images.IDFromName(os.client, os.config.Image)

	instanceName := fmt.Sprintf("krico.%s", uuid.New())

	serverOpts := servers.CreateOpts{
		Name:      instanceName,
		FlavorRef: flavorID,
		ImageRef:  image,
	}

	serverOptsExt := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           keypairName.Value(),
	}

	instance, err := servers.Create(os.client, serverOptsExt).Extract()

	if err != nil {
		err = errors.Wrapf(err, "%s Unable to create instance", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Scheduled instance %s creation", executorLogPrefix, instance.ID)

	err = servers.WaitForStatus(os.client, instance.ID, "ACTIVE", 60)

	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't launch instance", executorLogPrefix)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Launched instance %s", executorLogPrefix, instance.ID)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: floatingIP,
	}

	err = floatingips.AssociateInstance(os.client, instance.ID, associateOpts).ExtractErr()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't associate %q floating ip to instance %s", executorLogPrefix, floatingIP, instanceName)
		log.Error(err.Error())
		return nil, err
	}

	log.Infof("%s Associated %q floating ip", executorLogPrefix, floatingIP)

	remoteConfig := RemoteConfig{
		User:    os.config.User,
		KeyPath: os.config.SSHKeyPath,
		Port:    22,
	}

	// Wait while to ensure that everything booted up
	log.Infof("%s Waiting for %s instance to boot up", executorLogPrefix, instanceName)
	time.Sleep(bootUpTimeOut.Value())

	type serverAttributesExt struct {
		servers.Server
		extendedserverattributes.ServerAttributesExt
	}

	var extendedAttributes serverAttributesExt

	err = servers.Get(os.client, instance.ID).ExtractInto(&extendedAttributes)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't get %s instance extended attributes", executorLogPrefix, instanceName)
		log.Error(err)
		return nil, err
	}

	os.config.HypervisorInstanceName = extendedAttributes.InstanceName

	remote, err := NewRemote(floatingIP, remoteConfig)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't create remote executor for %s instance with %s ip", executorLogPrefix, instanceName, floatingIP)
		log.Error(err)
		return nil, err
	}

	log.Infof("%s Created remote executor", executorLogPrefix)

	remoteHandler, err := remote.Execute(command)
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't execute %q on %s (%s)", executorLogPrefix, command, instanceName, floatingIP)
		log.Error(err)
		return nil, err
	}

	exitCode, err := remoteHandler.ExitCode()
	if err != nil {
		err = errors.Wrapf(err, "%s Couldn't obtain exit code from executed command", executorLogPrefix)
	}

	log.Infof("%s Executed %q on %s (%s) with exit code: %d", executorLogPrefix, command, instanceName, floatingIP, exitCode)

	outputDirectory, err := createOutputDirectory(command, "openstack")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(outputDirectory)
	if err != nil {
		removeDirectory(outputDirectory)
		err = errors.Wrapf(err, "%s Cannot create output files for command: %q", executorLogPrefix, command)
		log.Error(err.Error())
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
		os:             &os,
		running:        true,
		exitCode:       exitCode,
		requestStop:    make(chan struct{}),
		requestDelete:  make(chan struct{}),
		stopped:        make(chan struct{}),
		deleted:        make(chan struct{}),
	}

	taskWatcher := &openstackWatcher{
		instance:      instance.ID,
		client:        os.client,
		running:       &taskHandle.running,
		requestStop:   taskHandle.requestStop,
		requestDelete: taskHandle.requestDelete,
		stopped:       taskHandle.stopped,
		deleted:       taskHandle.deleted,
	}

	err = taskWatcher.watch()

	return taskHandle, nil
}

func (os Openstack) ensureFlavor() (string, error) {
	os.config.FlavorName = fmt.Sprintf("krico.cpu-%d.ram-%d.disk-%d", os.config.Flavor.VCPUs, os.config.Flavor.RAM, os.config.Flavor.Disk)

	listOpts := flavors.ListOpts{
		AccessType: flavors.AllAccess,
	}

	allPages, err := flavors.ListDetail(os.client, listOpts).AllPages()
	if err != nil {
		err = errors.Wrapf(err, "Couldn't get flavors list to check %s within it", os.config.FlavorName)
		return "", err
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't extract flavors list to check %s within it", os.config.FlavorName)
		return "", err
	}

	// check if flavor already exists
	for _, flavor := range allFlavors {
		if flavor.Name == os.config.FlavorName {
			log.Infof("%s Using existing flavor %q", executorLogPrefix, os.config.FlavorName)
			return flavor.ID, nil
		}
	}

	// create flavor
	flavorID := uuid.New()
	flavorOpts := flavors.CreateOpts{
		ID:    flavorID,
		Name:  os.config.FlavorName,
		Disk:  &os.config.Flavor.Disk,
		RAM:   os.config.Flavor.RAM,
		VCPUs: os.config.Flavor.VCPUs,
	}
	_, err = flavors.Create(os.client, flavorOpts).Extract()
	if err != nil {
		err = errors.Wrapf(err, "Unable to create flavor %q", os.config.FlavorName)
		return flavorID, err
	}

	log.Infof("%s Created flavor %q", executorLogPrefix, os.config.FlavorName)

	return flavorID, nil
}

func (os Openstack) ensureKeypair() error {
	allPages, err := keypairs.List(os.client).AllPages()
	if err != nil {
		err = errors.Wrap(err, "Couldn't read keypairs list")
		return err
	}

	allKeyPairs, err := keypairs.ExtractKeyPairs(allPages)
	if err != nil {
		err = errors.Wrap(err, "Couldn't extract keypairs list")
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
	publicKeyPath := os.config.SSHKeyPath + ".pub"
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't read public key %q", publicKeyPath)
		return err
	}

	keypairOpts := keypairs.CreateOpts{
		Name:      keypairName.Value(),
		PublicKey: string(publicKey),
	}

	_, err = keypairs.Create(os.client, keypairOpts).Extract()
	if err != nil {
		err = errors.Wrapf(err, "Couldn't create keypair %q", keypairName.Value())
		return err
	}

	log.Infof("%s Created keypair %q", executorLogPrefix, keypairName.Value())

	return nil
}

func (os Openstack) findFloatingIP() (string, error) {
	allPages, err := floatingips.List(os.client).AllPages()
	if err != nil {
		err = errors.Wrap(err, "Couldn't read floating IPs list")
		return "", err
	}

	allFloatingIPs, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		err = errors.Wrap(err, "Couldn't extract floating IPs list")
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
		err = errors.New("Couldn't find free floating ip")
		return "", err
	}

	return floatingIP.IP, nil
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
