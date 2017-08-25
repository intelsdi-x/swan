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

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
	"github.com/pkg/errors"
)

// DefaultConfig returns default configuration for Caffe Inference Collector session.
func DefaultConfig() snap.SessionConfig {
	pub := publishers.NewDefaultPublisher()

	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			snap.CaffeInferenceCollector,
			pub.PluginName},
		TaskName: "swan-caffe-inference-session",
		Metrics: []string{
			"/intel/swan/caffe/inference/*/batches",
		},
	}
}

// CaffeSession configures & launches snap workflow for gathering
// SLIs from Caffe Inference.
type CaffeSession struct {
	session *snap.Session
}

// NewSessionLauncher constructs CaffeInferenceSnapSessionLauncher.
func NewSessionLauncher(config snap.SessionConfig) (*CaffeSession, error) {
	session, err := snap.NewSessionLauncher(config)

	if err != nil {
		return nil, err
	}
	return &CaffeSession{
		session: session,
	}, nil
}

// NewSessionLauncherDefault constructs CaffeSession.
func NewSessionLauncherDefault() (*CaffeSession, error) {
	session, err := snap.NewSessionLauncher(DefaultConfig())

	if err != nil {
		return nil, err
	}
	return &CaffeSession{
		session: session,
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
