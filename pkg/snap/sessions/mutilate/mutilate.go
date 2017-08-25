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

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
)

// DefaultConfig returns default configuration for Mutilate Collector session.
func DefaultConfig() snap.SessionConfig {
	pub := publishers.NewDefaultPublisher()
	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			snap.MutilateCollector,
			pub.PluginName},
		TaskName: "swan-mutilate-session",
		Metrics: []string{
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
	}
}

// MutilateSessionLauncher configures & launches snap workflow for gathering
// SLIs from Mutilate.
type MutilateSession struct {
	session *snap.Session
}

// NewSessionLauncherDefault creates SessionLauncher based on values
// returned by DefaultConfig().
func NewSessionLauncherDefault() (*MutilateSession, error) {
	session, err := snap.NewSessionLauncher(DefaultConfig())

	if err != nil {
		return nil, err
	}
	return &MutilateSession{
		session: session,
	}, nil
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *MutilateSession) LaunchSession(
	task executor.TaskInfo,
	tags map[string]interface{}) (executor.TaskHandle, error) {

	// Obtain Mutilate output file.
	stdoutFile, err := task.StdoutFile()
	if err != nil {
		return nil, err
	}

	// Configuring Mutilate collector.
	s.session.CollectNodeConfigItems = []snap.CollectNodeConfigItem{
		{
			Ns:    "/intel/swan/mutilate",
			Key:   "stdout_file",
			Value: stdoutFile.Name(),
		},
	}

	// Start session.
	handle, err := s.session.Launch(tags)
	if err != nil {
		return nil, err
	}

	return handle, nil
}
