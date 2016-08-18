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
	gopathLocation := os.Getenv("GOPATH")
	if os.Getenv("KUBERNETES") != "" {
		gopathLocation = "$GOPATH"
	}
	return path.Join(gopathLocation, "src", swanPkg)
}

// GetSwanBuildPath returns absolute path to Swan build directory.
func GetSwanBuildPath() string {
	return path.Join(GetSwanPath(), "build")
}

// GetSwanWorkloadsPath returns absolute path to Swan workloads directory.
func GetSwanWorkloadsPath() string {
	return path.Join(GetSwanPath(), "workloads")
}

// GetSwanExperimentPath returns absolute path to Swan experiment directory.
func GetSwanExperimentPath() string {
	return path.Join(GetSwanPath(), "experiments")
}

// GetSwanBinPath returns absolute path to Swan misc/bin directory.
func GetSwanBinPath() string {
	return path.Join(GetSwanPath(), "misc", "bin")
}
