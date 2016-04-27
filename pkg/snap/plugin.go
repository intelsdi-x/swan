package snap

import (
	"os"
	"path"
)

// LoadPlugin loads plugin binary. Currently searching the swan repo only.
func LoadPlugin(name string) error {
	err := ensureConnected()
	if err != nil {
		return err
	}

	// Current workaround to load plugins from swan repo
	// Test will run in pkg/swan to backing up two directories to get to 'build' directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	paths := []string{path.Join(cwd, "..", "..", "build", name)}

	r := pClient.LoadPlugin(paths)
	if r.Err != nil {
		return r.Err
	}

	return nil
}

// IsPluginLoaded connects to snap and looks for plugin with given name.
// Be aware that the name is not the plugin binary path or name but defined by the plugin itself.
func IsPluginLoaded(name string) (bool, error) {
	err := ensureConnected()
	if err != nil {
		return false, err
	}

	// Get all (running: false) plugins
	plugins := pClient.GetPlugins(false)
	if plugins.Err != nil {
		return false, plugins.Err
	}

	for _, lp := range plugins.LoadedPlugins {
		if name == lp.Name {
			return true, nil
		}
	}

	return false, nil
}

// UnloadPlugin unloads a plugin of type t and given name and version.
func UnloadPlugin(t string, name string, version int) error {
	err := ensureConnected()
	if err != nil {
		return err
	}

	r := pClient.UnloadPlugin(t, name, version)
	if r.Err != nil {
		return r.Err
	}

	return nil
}
