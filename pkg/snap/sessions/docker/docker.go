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

package docker

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
)

// DefaultConfig returns default configuration for Docker Session Launcher.
func DefaultConfig() snap.SessionConfig {
	pub := publishers.NewDefaultPublisher()

	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			snap.DockerCollector,
			pub.PluginName},
		TaskName: "swan-docker-session",
		Metrics: []string{
			"/intel/docker/*/stats/cgroups/*",
		},
	}
}

// TODO
type DockerSession struct {
	session *snap.Session
}

// NewSessionLauncher constructs Docker Session Launcher.
func NewSessionLauncher(config snap.SessionConfig) (*DockerSession, error) {
	session, err := snap.NewSessionLauncher(config)

	if err != nil {
		return nil, err
	}
	return &DockerSession{
		session: session,
	}, nil
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *DockerSession) Launch() (executor.TaskHandle, error) {
	// Start session.
	return s.session.Launch()
}

// String returns human readable name for job.
func (s *SessionLauncher) String() string {
	return "Snap Docker Collection"
}
