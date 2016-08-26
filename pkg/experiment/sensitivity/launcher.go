package sensitivity

import (
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
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
