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
	"os/exec"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/pkg/errors"
)

const (
	// CaffeInferenceCollector is a name of the snap plugin binary
	CaffeInferenceCollector = "snap-plugin-collector-caffe-inference"
	// DockerCollector is name of snap plugin binary.
	DockerCollector = "snap-plugin-collector-docker"
	// MutilateCollector is name of snap plugin binary.
	MutilateCollector string = "snap-plugin-collector-mutilate"
	// RDTCollector is Snap RDT Metric collector
	RDTCollector = "snap-plugin-collector-rdt"
	// SPECjbbCollector is name of snap plugin binary used to collect metrics from SPECjbb output file.
	SPECjbbCollector string = "snap-plugin-collector-specjbb"

	// CassandraPublisher is name of snap plugin binary.
	CassandraPublisher = "snap-plugin-publisher-cassandra"
	// FilePublisher is Snap testing file publisher
	FilePublisher = "snap-plugin-publisher-file"
	// SessionPublisher is name of snap plugin binary.
	SessionPublisher = "snap-plugin-publisher-session-test"
)

var (
	// SnapteldAddress represents snap daemon address flag.
	SnapteldAddress = conf.NewStringFlag("snapteld_address", "Snapteld address in `http://%s:%s` format", "http://127.0.0.1:8181")
)

// DefaultPluginLoaderConfig returns default config for PluginLoader.
func DefaultPluginLoaderConfig() PluginLoaderConfig {
	return PluginLoaderConfig{
		SnapteldAddress: SnapteldAddress.Value(),
	}
}

// PluginLoaderConfig contains configuration for PluginLoader.
type PluginLoaderConfig struct {
	SnapteldAddress string
}

// PluginLoader is used to simplify Snap plugin loading.
type PluginLoader struct {
	pluginsClient *Plugins
	config        PluginLoaderConfig
}

// NewDefaultPluginLoader returns PluginLoader with DefaultPluginLoaderConfig.
// Returns error when could not connect to Snap.
func NewDefaultPluginLoader() (*PluginLoader, error) {
	return NewPluginLoader(DefaultPluginLoaderConfig())
}

// NewPluginLoader constructs PluginLoader with given config.
// Returns error when could not connect to Snap.
func NewPluginLoader(config PluginLoaderConfig) (*PluginLoader, error) {
	snapClient, err := client.New(config.SnapteldAddress, "v1", true)
	if err != nil {
		return nil, err
	}
	plugins := NewPlugins(snapClient)

	return &PluginLoader{
		pluginsClient: plugins,
		config:        config,
	}, nil
}

// Load loads supplied plugin names from plugin path and returns slice of
// encountered errors.
func (l PluginLoader) Load(plugins ...string) error {
	var errors errcollection.ErrorCollection
	for _, plugin := range plugins {
		err := l.load(plugin)
		errors.Add(err)
	}
	return errors.GetErrIfAny()
}

// load loads selected plugin from plugin path.
func (l PluginLoader) load(plugin string) error {
	pluginName, pluginType, err := GetPluginNameAndType(plugin)
	if err != nil {
		return err
	}

	isPluginLoaded, err := l.pluginsClient.IsLoaded(pluginType, pluginName)
	if err != nil {
		return err
	}
	if isPluginLoaded {
		return nil
	}

	pluginPath, err := exec.LookPath(plugin)
	if err != nil {
		return errors.Wrapf(err, "cannot find snap plugin %s in $PATH", plugin)
	}

	return l.pluginsClient.LoadPlugin(pluginPath)
}
