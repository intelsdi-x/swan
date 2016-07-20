package snap

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/pkg/errors"
)

// Plugins provides a 'manager' like abstraction for plugin operations.
type Plugins struct {
	// Client to Snapd
	pClient *client.Client
}

// NewPlugins returns an instantiated Plugins object, using pClient for all plugin operations.
func NewPlugins(pClient *client.Client) *Plugins {
	return &Plugins{
		pClient: pClient,
	}
}

// LoadPlugin loads plugin binary.
func (p *Plugins) LoadPlugin(pluginPath string) error {
	return p.LoadPlugins([]string{pluginPath})
}

// LoadPlugins loads multiple snap plugins.
func (p *Plugins) LoadPlugins(pluginPaths []string) error {
	r := p.pClient.LoadPlugin(pluginPaths)
	if r.Err != nil {
		return errors.Wrapf(r.Err, "could not load plugin from path %q", pluginPaths)
	}

	return nil
}

// IsLoaded connects to snap and looks for plugin with given name.
// Be aware that the name is not the plugin binary path or name but defined by the plugin itself.
func (p *Plugins) IsLoaded(t string, name string) (bool, error) {
	// Get all (running: false) plugins
	plugins := p.pClient.GetPlugins(false)
	if plugins.Err != nil {
		return false, errors.Wrap(plugins.Err, "could not obtain loaded plugins")
	}

	for _, lp := range plugins.LoadedPlugins {
		if t == lp.Type && name == lp.Name {
			return true, nil
		}
	}

	return false, nil
}

// Unload unloads a plugin of type t and given name and version.
func (p *Plugins) Unload(t string, name string, version int) error {
	r := p.pClient.UnloadPlugin(t, name, version)
	if r.Err != nil {
		return errors.Wrapf(r.Err, "could not unload plugin %q:%q:%q",
			name, t, version)
	}

	return nil
}
