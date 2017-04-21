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

package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
)

// LauncherSessionPair is a pair of Launcher and corresponding Session Launcher.
// TODO(bp): We can think about moving to unified Launcher which launch both
// Launcher and SessionLauncher. It is not possible right now since Launcher
// and LauncherSession have not the same API.
type LauncherSessionPair struct {
	Launcher            executor.Launcher
	SnapSessionLauncher snap.SessionLauncher
}

// NewLauncherWithoutSession constructs LauncherSessionPair without any Session.
func NewLauncherWithoutSession(launcher executor.Launcher) LauncherSessionPair {
	return LauncherSessionPair{launcher, nil}
}

// NewMonitoredLauncher constructs LauncherSessionPair with specified Session.
func NewMonitoredLauncher(
	launcher executor.Launcher,
	snapSessionLauncher snap.SessionLauncher) LauncherSessionPair {
	return LauncherSessionPair{launcher, snapSessionLauncher}
}

// LoadGeneratorSessionPair is a pair of Load Generator and corresponding Session Launcher.
type LoadGeneratorSessionPair struct {
	LoadGenerator       executor.LoadGenerator
	SnapSessionLauncher snap.SessionLauncher
}

// NewLoadGeneratorWithoutSession constructs LoadGenerator without any Session.
func NewLoadGeneratorWithoutSession(
	loadGenerator executor.LoadGenerator) LoadGeneratorSessionPair {

	return LoadGeneratorSessionPair{loadGenerator, nil}
}

// NewMonitoredLoadGenerator constructs LoadGenerator with specified Session.
func NewMonitoredLoadGenerator(
	loadGenerator executor.LoadGenerator,
	snapSessionLauncher snap.SessionLauncher) LoadGeneratorSessionPair {

	return LoadGeneratorSessionPair{loadGenerator, snapSessionLauncher}
}
