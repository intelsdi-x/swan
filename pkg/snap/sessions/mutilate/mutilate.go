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

package mutilatesession

import (
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
)

// DefaultConfig returns default configuration for Mutilate Collector session.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", snap.PluginAnyVersion)
	sessions.ApplyInfluxDBConfiguration(publisher)

	return Config{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       publisher,
	}
}

// DefaultConfig returns default configuration for Mutilate Collector session.
func DefaultInfluxDBConfig() Config {
	publisher := wmap.NewPublishNode("influxdb", snap.PluginAnyVersion)
	sessions.ApplyInfluxDBConfiguration(publisher)

	return Config{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       publisher,
	}
}

// Config contains configuration for Mutilate Collector session.
type Config struct {
	SnapteldAddress string
	Publisher       *wmap.PublishWorkflowMapNode
	Interval        time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering
// SLIs from Mutilate.
type SessionLauncher struct {
	session                *snap.Session
	snapClient             *client.Client
	mutilateOutputFilePath string
}

// NewSessionLauncherDefault creates SessionLauncher based on values
// returned by DefaultConfig().
func NewSessionLauncherDefault(
	mutilateOutputFilePath string,
	tags map[string]interface{}) (*SessionLauncher, error) {
	return NewSessionLauncher(mutilateOutputFilePath, tags, DefaultConfig())
}

// NewSessionLauncher constructs MutilateSnapSessionLauncher.
func NewSessionLauncher(
	mutilateOutputFilePath string,
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
	err = loader.Load(snap.MutilateCollector, snap.InfluxDBPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession(
			"swan-mutilate-session",
			[]string{
				"/intel/swan/mutilate/*/avg",
				"/intel/swan/mutilate/*/std",
				"/intel/swan/mutilate/*/min",
				"/intel/swan/mutilate/*/percentile/5th",
				"/intel/swan/mutilate/*/percentile/10th",
				"/intel/swan/mutilate/*/percentile/90th",
				"/intel/swan/mutilate/*/percentile/95th",
				"/intel/swan/mutilate/*/percentile/99th",
				"/intel/swan/mutilate/*/qps",
			},
			config.Interval,
			snapClient,
			config.Publisher,
			tags,
		),
		snapClient:             snapClient,
		mutilateOutputFilePath: mutilateOutputFilePath,
	}, nil
}

// Launch starts Snap Collection session and returns handle to that session.
func (s *SessionLauncher) Launch() (executor.TaskHandle, error) {
	// Configuring Mutilate collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		{
			Ns:    "/intel/swan/mutilate",
			Key:   "stdout_file",
			Value: s.mutilateOutputFilePath,
		},
	}

	return s.session.Launch()
}

// String returns human readable name for job.
func (s *SessionLauncher) String() string {
	return "Snap Mutilate Collection"
}
