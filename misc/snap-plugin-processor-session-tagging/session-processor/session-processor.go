package sessionProcessor

import (
	"bytes"
	"encoding/gob"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name       = "session-processor"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

// The SessionProcessor adds experiment and phase tags to all incoming metrics.
// This enables the publishers to store metrics according to the running swan experiment and session.
type SessionProcessor struct{}

// Process is an implementation needed for the Processor interface and here,
// adds swan_experiment and swan_phase tags to all metrics.
func (p *SessionProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	var metrics []plugin.MetricType

	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		return "", nil, err
	}

	swanExperiment := config["swan_experiment"].(ctypes.ConfigValueStr).Value
	swanPhase := config["swan_phase"].(ctypes.ConfigValueStr).Value
	swanRepetition := config["swan_repetition"].(ctypes.ConfigValueStr).Value

	for idx := range metrics {
		metrics[idx].Tags_ = map[string]string{
			"swan_experiment": swanExperiment,
			"swan_phase":      swanPhase,
			"swan_repetition": swanRepetition,
		}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)
	content = buf.Bytes()

	return contentType, content, nil
}

// GetConfigPolicy is an implementation needed for the Processor interface and here,
// returns configuration requiring 'swan_experiment' and 'swan_phase' to be set.
func (p *SessionProcessor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("swan_experiment", true)
	if err != nil {
		panic(err)
	}
	r1.Description = "Swan experiment ID to tag metrics with"
	config.Add(r1)

	r2, err := cpolicy.NewStringRule("swan_phase", true)
	if err != nil {
		panic(err)
	}
	r1.Description = "Swan phase ID to tag metrics with"
	config.Add(r2)

	cp.Add([]string{""}, config)
	return cp, nil
}

// Meta returns a plugin meta data.
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}
