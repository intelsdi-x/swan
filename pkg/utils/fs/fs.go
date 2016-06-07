package fs

import (
	"os"
	"path"
)

const (
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
