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

package specjbbsession

import (
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
)

// DefaultConfig returns default configuration for SPECjbb Collector session.
func DefaultConfig() Config {
	publisher := wmap.NewPublishNode("cassandra", snap.PluginAnyVersion)
	sessions.ApplyCassandraConfiguration(publisher)

	return Config{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       publisher,
	}
}

// Config contains configuration for SPECjbb Collector session.
type Config struct {
	SnapteldAddress string
	Publisher       *wmap.PublishWorkflowMapNode
	Interval        time.Duration
}

// SessionLauncher configures & launches snap workflow for gathering
// metrics from SPECjbb.
type SessionLauncher struct {
	session    *snap.Session
	snapClient *client.Client

	specjbbOutputFilePath string
}

// NewSessionLauncherDefault creates SessionLauncher based on values
// returned by DefaultConfig().
func NewSessionLauncherDefault(
	specjbbOutputFilePath string,
	tags map[string]interface{}) (*SessionLauncher, error) {
	return NewSessionLauncher(specjbbOutputFilePath, tags, DefaultConfig())
}

// NewSessionLauncher constructs SPECjbbSnapSessionLauncher.
func NewSessionLauncher(
	specjbbOutputFilePath string,
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

	err = loader.Load(snap.SPECjbbCollector, snap.CassandraPublisher)
	if err != nil {
		return nil, err
	}

	return &SessionLauncher{
		session: snap.NewSession(
			"swan-specjbb-session",
			[]string{
				"/intel/swan/specjbb/*/min",
				"/intel/swan/specjbb/*/percentile/50th",
				"/intel/swan/specjbb/*/percentile/90th",
				"/intel/swan/specjbb/*/percentile/95th",
				"/intel/swan/specjbb/*/percentile/99th",
				"/intel/swan/specjbb/*/max",
				"/intel/swan/specjbb/*/qps",
				"/intel/swan/specjbb/*/issued_requests",
			},
			config.Interval,
			snapClient,
			config.Publisher,
			tags,
		),
		snapClient:            snapClient,
		specjbbOutputFilePath: specjbbOutputFilePath,
	}, nil
}

// Launch starts Snap Collection session and returns handle to that session.
func (s *SessionLauncher) Launch() (executor.TaskHandle, error) {
	// Configuring SPECjbb collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		{
			Ns:    "/intel/swan/specjbb",
			Key:   "stdout_file",
			Value: s.specjbbOutputFilePath,
		},
	}

	return s.session.Launch()
}

// String returns human readable name for job.
func (s *SessionLauncher) String() string {
	return "Snap specJBB Collection"
}
