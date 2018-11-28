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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/extendedserverattributes"
)

const (
	kricoKeypairName = "krico"
	bootUpTimeOut    = 30
)

var (
	flavorDiskFlag  = conf.NewIntFlag("flavor_disk", "Openstack flavor disk size [GB]", 10)
	flavorRAMFlag   = conf.NewIntFlag("flavor_ram", "Openstack flavor RAM size [MB]", 1024)
	flavorVCPUsFlag = conf.NewIntFlag("flavor_vcpus", "Openstack flavor VCPUs", 1)
	imageFlag       = conf.NewStringFlag("image", "Name of image.", "cirros")
	userFlag        = conf.NewStringFlag("username", "Username", "cirros")
	SSHKeyPathFlag  = conf.NewStringFlag("ssh_key", "SSH key path", "~/.ssh/id_rsa")
)

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
		SSHKeyPath: SSHKeyPathFlag.Value(),
	}
}

// OpenstackFlavor defines Openstack flavor data.
type OpenstackFlavor struct {
	Disk  int
	RAM   int
	VCPUs int
}

// OpenstackAuthConfig defines authentication configuration.
type OpenstackAuthConfig struct {
	Username   string
	Password   string
	TenantID   string
	DomainName string
	Endpoint   string
}

// OpenstackConfig defines Openstack configuration.
type OpenstackConfig struct {
	Auth       gophercloud.AuthOptions
	Flavor     OpenstackFlavor
	FlavorName string
	Image      string
	User       string
	SSHKeyPath string
	HypervisorInstanceName string
}

// Openstack defines server configuration and client.
type Openstack struct {
	config *OpenstackConfig
	client *gophercloud.ServiceClient
	instanceCgroup string
}

// NewOpenstack creating new Openstack object.
func NewOpenstack(config *OpenstackConfig) Executor {
	return Openstack{config: config}
}

func (os Openstack) String() string {
	return fmt.Sprintf("Openstack Executor at %s", os.config.Auth.IdentityEndpoint)
}

// Execute creates a instance and runs the provided command in it. When the command completes, the instance
// is stopped i.e. the container is not restarted automatically.
func (os Openstack) Execute(command string) (TaskHandle, error) {
	provider, err := openstack.AuthenticatedClient(os.config.Auth)
	if err != nil {
		return nil, fmt.Errorf("couldn't get provider: %v", err)
	}

	os.client, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	if err != nil {
		return nil, fmt.Errorf("couldn't get compute client: %v", err)
	}

	flavorID, err := os.ensureFlavor()
	if err != nil {
		return nil, err
	}

	if err = os.ensureKeypair(); err != nil {
		return nil, err
	}

	floatingIP, err := os.findFloatingIP()
	if err != nil {
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
		KeyName:           kricoKeypairName,
	}

	instance, err := servers.Create(os.client, serverOptsExt).Extract()
	if err != nil {
		return nil, fmt.Errorf("unable to create instance: %v", err)
	}
	log.Debugf("Scheduled instance %s creation", instance.ID)

	err = servers.WaitForStatus(os.client, instance.ID, "ACTIVE", 60)
	if err != nil {
		return nil, fmt.Errorf("couldn't launch instance: %v", err)
	}
	log.Debugf("Launched instance %s", instance.ID)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: floatingIP,
	}

	err = floatingips.AssociateInstance(os.client, instance.ID, associateOpts).ExtractErr()
	if err != nil {
		return nil, fmt.Errorf("couldn't associate floating ip to instance: %v", err)
	}

	log.Debugf("Associated %q floating ip", floatingIP)

	remoteConfig := RemoteConfig{
		User:    os.config.User,
		KeyPath: os.config.SSHKeyPath,
		Port:    22,
	}

	log.Debug("Waiting for instance to boot up")

	// Wait while to ensure that everything booted up!
	time.Sleep(time.Second*bootUpTimeOut)

	type serverAttributesExt struct {
		servers.Server
		extendedserverattributes.ServerAttributesExt
	}

	var extendedAttributes serverAttributesExt

	err = servers.Get(os.client, instance.ID).ExtractInto(&extendedAttributes)
	if err != nil {
		return nil, fmt.Errorf("couldn't get instance extended attributes: %v", err)
	}

	os.config.HypervisorInstanceName = extendedAttributes.InstanceName

	remote, err := NewRemote(floatingIP, remoteConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't create remote executor: %v", err)
	}
	log.Debug("Created remote executor")

	_, err = remote.Execute(command)
	if err != nil {
		return nil, fmt.Errorf("couldn't start remote task: %v", err)
	}

	outputDirectory, err := createOutputDirectory(command, "openstack")
	if err != nil {
		log.Errorf("cannot create output directory for command %q: %s", command, err.Error())
		return nil, err
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(outputDirectory)
	if err != nil {
		removeDirectory(outputDirectory)
		log.Errorf("cannot create output files for command %q: %s", command, err.Error())
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

func (os Openstack) GetInstanceCgroup() (string, error){

	if os.instanceCgroup == "" {
		return "", fmt.Errorf("No instance cgroup! Probably instance not work!")
	}

	return os.instanceCgroup, nil
}

func (os Openstack) ensureFlavor() (string, error) {
	os.config.FlavorName = fmt.Sprintf("krico.cpu-%d.ram-%d.disk-%d", os.config.Flavor.VCPUs, os.config.Flavor.RAM, os.config.Flavor.Disk)

	listOpts := flavors.ListOpts{
		AccessType: flavors.AllAccess,
	}

	allPages, err := flavors.ListDetail(os.client, listOpts).AllPages()
	if err != nil {
		return "", fmt.Errorf("couldn't get flavors list: %v", err)
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", fmt.Errorf("couldn't extract flavors list: %v", err)
	}

	// check if flavor already exists
	for _, flavor := range allFlavors {
		if flavor.Name == os.config.FlavorName {
			log.Debug("Using existing flavors")
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
		return flavorID, fmt.Errorf("unable to create flavor: %v", err)
	}

	log.Debug("Created flavor")

	return flavorID, nil
}

func (os Openstack) ensureKeypair() error {
	allPages, err := keypairs.List(os.client).AllPages()
	if err != nil {
		return fmt.Errorf("couldn't read keypairs list: %v", err)
	}

	allKeyPairs, err := keypairs.ExtractKeyPairs(allPages)
	if err != nil {
		return fmt.Errorf("couldn't extract keypairs list: %v", err)
	}

	// check if keypair already exists
	for _, kp := range allKeyPairs {
		if kp.Name == kricoKeypairName {
			log.Debug("Using exists keypair")
			return nil
		}
	}

	// create keypair
	publicKey, err := ioutil.ReadFile(os.config.SSHKeyPath + ".pub")
	if err != nil {
		return fmt.Errorf("couldn't read public key: %s", err)
	}

	keypairOpts := keypairs.CreateOpts{
		Name:      kricoKeypairName,
		PublicKey: string(publicKey),
	}

	_, err = keypairs.Create(os.client, keypairOpts).Extract()
	if err != nil {
		return fmt.Errorf("couldn't create keypair: %v", err)
	}

	log.Debug("Created keypair")

	return nil
}

func (os Openstack) findFloatingIP() (string, error) {
	allPages, err := floatingips.List(os.client).AllPages()
	if err != nil {
		return "", fmt.Errorf("couldn't read floatingips list: %v", err)
	}

	allFloatingIPs, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return "", fmt.Errorf("couldn't extract floatingips lis: %v", err)
	}

	var floatingIP *floatingips.FloatingIP

	for _, fip := range allFloatingIPs {
		if fip.FixedIP == "" {
			floatingIP = &fip
			break
		}
	}

	if floatingIP == nil {
		return "", fmt.Errorf("couldn't find free floating ip")
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
	requestStop    chan struct{}
	requestDelete  chan struct{}
	stopped        chan struct{}
	deleted        chan struct{}
}

func (th *OpenstackTaskHandle) String() string {
	return fmt.Sprintf("OpenStack instance running with command: %q", th.command)
}

// Address returns address where task was located.
func (th *OpenstackTaskHandle) Address() string {
	return th.hostIP
}

// EraseOutput deletes the directory where output files resides.
func (th *OpenstackTaskHandle) EraseOutput() error {
	outputDir := filepath.Dir(th.stdoutFilePath)
	return removeDirectory(outputDir)
}

// ExitCode returns an exit code of finished task.
// Returns error if If task is not terminated.
func (th *OpenstackTaskHandle) ExitCode() (int, error) {

	if th.isRunning() {
		return 1, errors.New("task is still running")
	}

	return 0, nil
}

// Status returns a state of the task.
func (th *OpenstackTaskHandle) Status() TaskState {
	if th.isRunning() {
		return RUNNING
	}

	return TERMINATED
}

// StderrFile returns a file handle for file to the task's stderr file.
func (th *OpenstackTaskHandle) StderrFile() (*os.File, error) {
	return openFile(th.stderrFilePath)
}

// StdoutFile returns a file handle for file to the task's stdout file.
func (th *OpenstackTaskHandle) StdoutFile() (*os.File, error) {
	return openFile(th.stdoutFilePath)
}

// Stop stops a task.
// Returns error if something wrong has happen during task execution.
func (th *OpenstackTaskHandle) Stop() error {
	if !th.isRunning() {
		return nil
	}

	log.Debugf("Openstack task handle: delete instance %q", th.instance)

	th.requestStop <- struct{}{}

	log.Debugf("Openstack task handle: waiting for instance %q to stop", th.instance)

	<-th.stopped

	log.Debugf("Openstack task handle: instance %q stop", th.instance)

	th.requestDelete <- struct{}{}

	<-th.deleted

	log.Debugf("Openstack task handle: instance %q deleted", th.instance)

	return nil
}

// Wait blocks and waits for task to terminate.
// Parameter `timeout` is waiting timeout. For `0` it wil wait until task termination.
// Returns `terminated` true when task terminates.
// Returns error if something wrong has happen during task execution.
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
			case "SUSPENDED":
				*watcher.running = false
			case "RESCUED":
				*watcher.running = true
			case "STOPPED":
				*watcher.running = false
			case "SOFT_DELETED":
				*watcher.running = false
				watcher.requestDelete <- struct{}{}
			}
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
