package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/workloads"
)

// LauncherWithCollection is a pair of Launcher and corresponding Collection Launcher.
type LauncherWithCollection struct {
	Launcher           workloads.Launcher
	CollectionLauncher snap.SessionLauncher
}

// NewLauncher constructs Launcher without any Collection.
func NewLauncher(launcher workloads.Launcher) LauncherWithCollection {
	return LauncherWithCollection{launcher, nil}
}

// NewLauncherWithCollection constructs Launcher with specified Collection.
func NewLauncherWithCollection(
	launcher workloads.Launcher,
	collectionLauncher snap.SessionLauncher) LauncherWithCollection {
	return LauncherWithCollection{launcher, collectionLauncher}
}

// LoadGeneratorWithCollection is a pair of Load Generator and corresponding Collection Launcher.
type LoadGeneratorWithCollection struct {
	LoadGenerator      workloads.LoadGenerator
	CollectionLauncher snap.SessionLauncher
}

// NewLoadGenerator constructs LoadGenerator without any Collection.
func NewLoadGenerator(
	loadGenerator workloads.LoadGenerator) LoadGeneratorWithCollection {

	return LoadGeneratorWithCollection{loadGenerator, nil}
}

// NewLoadGeneratorWithCollection constructs LoadGenerator with specified Collection.
func NewLoadGeneratorWithCollection(
	loadGenerator workloads.LoadGenerator,
	collectionLauncher snap.SessionLauncher) LoadGeneratorWithCollection {

	return LoadGeneratorWithCollection{loadGenerator, collectionLauncher}
}
