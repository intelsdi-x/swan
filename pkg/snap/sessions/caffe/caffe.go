// Copyright (c) 2017 Intel Corporation
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

package caffeinferencesession

import (
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/pkg/errors"
)

// DefaultConfig returns default configuration for Caffe Inference Collector session.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", snap.PluginAnyVersion)
	sessions.ApplyCassandraConfiguration(publisher)

	return Config{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       publisher,
	}
}

// Config contains configuration for Caffe Inference Collector session.
type Config struct {
	SnapteldAddress string
	Publisher       *wmap.PublishWorkflowMapNode
	Interval        time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering
// SLIs from Caffe Inference.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client

	caffe executor.Launcher
}

//NewDefaultSessionLauncher constructs CaffeInferenceSnapSessionLauncher with default config.
func NewDefaultSessionLauncher(caffe executor.Launcher, tags map[string]interface{}) (*SessionLauncher, error) {
	return NewSessionLauncher(caffe, tags, DefaultConfig())
}

// NewSessionLauncher constructs CaffeInferenceSnapSessionLauncher.
func NewSessionLauncher(
	caffe executor.Launcher,
	tags map[string]interface{},
	config Config) (*SessionLauncher, error) {
	snapClient, err := client.New(config.SnapteldAddress, "v1", true)
	if err != nil {
		return nil, err
	}

	loaderConfig := snap.DefaultPluginLoaderConfig()
	loaderConfig.SnapteldAddress = config.SnapteldAddress
	loader, err := snap.NewPluginLoader(loaderConfig)
	if err != nil {
		return nil, err
	}

	err = loader.Load(snap.CaffeInferenceCollector, snap.CassandraPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession(
			"swan-caffe-inference-session",
			[]string{"/intel/swan/caffe/inference/*/batches"},
			config.Interval,
			snapClient,
			config.Publisher,
			tags,
		),
		snapClient: snapClient,
		caffe:      caffe,
	}, nil
}

// Launch starts Snap Collection session and returns handle to that session.
func (s *SessionLauncher) Launch() (executor.TaskHandle, error) {
	caffeHandle, err := s.caffe.Launch()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot launch Caffe workload")
	}

	stdout, err := caffeHandle.StdoutFile()
	if err != nil {
		caffeHandle.Stop()
		return nil, errors.Wrapf(err, "cannot get Caffe stdout file for metrics collection")
	}
	defer stdout.Close()

	// Configuring Caffe collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		{
			Ns:    "/intel/swan/caffe/inference",
			Key:   "stdout_file",
			Value: stdout.Name(),
		},
	}

	// Start metrics collection.
	collectionHandle, err := s.session.Launch()
	if err != nil {
		caffeHandle.Stop()
		return nil, errors.Wrapf(err, "cannot launch Caffe snap metrics collection")
	}

	return executor.NewClusterTaskHandle(caffeHandle, []executor.TaskHandle{collectionHandle}), nil
}

// String returns human readable name for job.
func (s *SessionLauncher) String() string {
	return "Snap Caffe Collection"
}
