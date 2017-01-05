package fs

import (
	"os"
	"path"
)

const (
	//TODO: refactor this; it needs to be project agnostic
	swanPkg = "github.com/intelsdi-x/swan"
)

// GetSwanPath returns absolute path to Swan root directory.
func GetSwanPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", swanPkg)
}

// GetSwanBuildPath returns absolute path to Swan build directory.
func GetSwanBuildPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", swanPkg, "build")
}

// GetSwanWorkloadsPath returns absolute path to Swan workloads directory.
func GetSwanWorkloadsPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", swanPkg, "workloads")
}

// GetSwanExperimentPath returns absolute path to Swan experiment directory.
func GetSwanExperimentPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", swanPkg, "experiments")
}

// GetSwanBinPath returns absolute path to Swan misc/bin directory.
func GetSwanBinPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", swanPkg, "misc", "bin")
}

const (
	athenaPkg = "github.com/intelsdi-x/athena"
)

// GetAthenaPath returns absolute path to Athena root directory.
func GetAthenaPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", athenaPkg)
}

// GetAthenaBuildPath returns absolute path to Athena build directory.
func GetAthenaBuildPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", athenaPkg, "build")
}

// GetAthenaBinPath returns absolute path to Athena misc/bin directory.
func GetAthenaBinPath() string {
	return path.Join(os.Getenv("GOPATH"), "src", athenaPkg, "misc", "bin")
}
