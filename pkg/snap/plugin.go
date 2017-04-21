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

package snap

import (
	"regexp"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/pkg/errors"
)

const (
	// PluginAnyVersion identifies any version of plugin that can be passed to Snap API.
	PluginAnyVersion = -1
)

// Plugins provides a 'manager' like abstraction for plugin operations.
type Plugins struct {
	// Client to Snapteld
	pClient *client.Client
}

// NewPlugins returns an instantiated Plugins object, using pClient for all plugin operations.
func NewPlugins(pClient *client.Client) *Plugins {
	return &Plugins{
		pClient: pClient,
	}
}

// LoadPlugin loads plugin binary from path.
func (p *Plugins) LoadPlugin(pluginPath string) error {
	r := p.pClient.LoadPlugin([]string{pluginPath})
	if r.Err != nil {
		return errors.Wrapf(r.Err, "could not load plugin from %q", pluginPath)
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

var (
	// TODO: remove optional (-plugin) group, after all plugins follow new naming scheme.
	pluginNameRegex      = regexp.MustCompile(".*?snap(-plugin)?.([a-z]+)-(.+)")
	pluginTypeGroupIndex = 2 // 0 is whole match, 1 is optional (-plugin) group.
	nameGroupIndex       = 3
)

// GetPluginNameAndType takes plugin binary name like "snap-plugin-collector-mutilate"
// and returns it's name and type inferred from binary name.
// Name must conform convention "*snap-plugin-(type)-(name)"
func GetPluginNameAndType(filename string) (name string, pluginType string, err error) {
	matches := pluginNameRegex.FindStringSubmatch(filename)

	if len(matches) != 4 {
		return "", "", errors.Errorf("GetPluginNameAndType regex failed on %q", filename)
	}

	name = matches[nameGroupIndex]
	pluginType = matches[pluginTypeGroupIndex]

	return name, pluginType, err
}
