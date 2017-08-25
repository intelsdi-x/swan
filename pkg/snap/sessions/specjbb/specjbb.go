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

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
)

// DefaultConfig returns default configuration for SPECjbb Collector session.
func DefaultConfig() snap.SessionConfig {
	pub := publishers.NewDefaultPublisher()

	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			snap.SPECjbbCollector,
			pub.PluginName},
		TaskName: "swan-specjbb-session",
		Metrics: []string{
			"/intel/swan/specjbb/*/min",
			"/intel/swan/specjbb/*/percentile/50th",
			"/intel/swan/specjbb/*/percentile/90th",
			"/intel/swan/specjbb/*/percentile/95th",
			"/intel/swan/specjbb/*/percentile/99th",
			"/intel/swan/specjbb/*/max",
			"/intel/swan/specjbb/*/qps",
			"/intel/swan/specjbb/*/issued_requests",
		},
	}
}

// SPECjbbSession configures & launches snap workflow for gathering
// metrics from SPECjbb.
type SPECjbbSession struct {
	session *snap.Session
}

// NewSessionLauncherDefault creates SessionLauncher based on values
// returned by DefaultConfig().
func NewSessionLauncherDefault() (*SPECjbbSession, error) {
	session, err := snap.NewSessionLauncher(DefaultConfig())

	if err != nil {
		return nil, err
	}
	return &SPECjbbSession{
		session: session,
	}, nil
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *SPECjbbSession) LaunchSession(
	task executor.TaskInfo,
	tags map[string]interface{}) (executor.TaskHandle, error) {

	// Obtain SPECjbb output file.
	stdoutFile, err := task.StdoutFile()
	if err != nil {
		return nil, err
	}

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
