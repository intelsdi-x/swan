package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/workloads"
)

// LauncherAndSessionPair is a pair of Launcher and corresponding Session Launcher.
// TODO(bp): We can think about moving to unified Launcher which launch both
// Launcher and SessionLauncher. It is not possible right now since Launcher
// and LauncherSession have not the same API.
type LauncherAndSessionPair struct {
	Launcher        workloads.Launcher
	SessionLauncher snap.SessionLauncher
}

// NewLauncherWithoutSession constructs LauncherAndSessionPair without any Session.
func NewLauncherWithoutSession(launcher workloads.Launcher) LauncherAndSessionPair {
	return LauncherAndSessionPair{launcher, nil}
}

// NewCollectedLauncher constructs WorkloadAndSessionPair with specified Session.
func NewCollectedLauncher(
	launcher workloads.Launcher,
	sessionLauncher snap.SessionLauncher) LauncherAndSessionPair {
	return LauncherAndSessionPair{launcher, sessionLauncher}
}

// LoadGeneratorAndSessionPair is a pair of Load Generator and corresponding Session Launcher.
type LoadGeneratorAndSessionPair struct {
	LoadGenerator   workloads.LoadGenerator
	SessionLauncher snap.SessionLauncher
}

// NewLoadGeneratorWithoutSession constructs LoadGenerator without any Session.
func NewLoadGeneratorWithoutSession(
	loadGenerator workloads.LoadGenerator) LoadGeneratorAndSessionPair {

	return LoadGeneratorAndSessionPair{loadGenerator, nil}
}

// NewCollectedLoadGenerator constructs LoadGenerator with specified Session.
func NewCollectedLoadGenerator(
	loadGenerator workloads.LoadGenerator,
	sessionLauncher snap.SessionLauncher) LoadGeneratorAndSessionPair {

	return LoadGeneratorAndSessionPair{loadGenerator, sessionLauncher}
}
