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

package use

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
)

// DefaultConfig returns default configuration for USE Collector session.
func DefaultConfig() snap.SessionConfig {
	pub := publishers.NewDefaultPublisher()

	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			snap.USECollector,
			pub.PluginName},
		TaskName: "swan-use-session",
		Metrics: []string{
			"/intel/use/compute/*",
			"/intel/use/memory/*",
		},
	}
}

// USESession configures & launches snap workflow for gathering
// metrics from USE.
type USESession struct {
	session *snap.Session
}

// NewSessionLauncherDefault creates SessionLauncher based on values
// returned by DefaultConfig().
func NewSessionLauncherDefault(config snap.SessionConfig) (*USESession, error) {
	session, err := snap.NewSessionLauncher(config)

	if err != nil {
		return nil, err
	}
	return &USESession{
		session: session,
	}, nil
}

// LaunchSession starts Snap Collection session and returns handle to that session.
func (s *USESession) Launch() (executor.TaskHandle, error) {

	// Start session.
	return s.session.Launch(tags)
}

// String returns human readable name for job.
func (s *SessionLauncher) String() string {
	return "Snap USE Collection"
}
