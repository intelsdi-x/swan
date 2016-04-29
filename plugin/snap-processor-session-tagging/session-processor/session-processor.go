package sessionProcessor

import (
	"bytes"
	"encoding/gob"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name       = "session-processor"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

type SessionProcessor struct{}

func (p *SessionProcessor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	return cp, nil
}

func (p *SessionProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	logger := log.New()

	var metrics []plugin.PluginMetricType

	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for idx, _ := range metrics {
		labels := metrics[idx].Labels_
		labels = append(labels, core.Label{Name: "swan-was-here"})
		metrics[idx].Labels_ = labels

		logger.Printf("Passed on metric: %v\n", metrics[idx])
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)
	content = buf.Bytes()

	return contentType, content, nil
}
