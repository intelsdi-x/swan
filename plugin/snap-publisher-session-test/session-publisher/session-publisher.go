package sessionPublisher

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name       = "session-test"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

type SessionPublisher struct {
}

func (f *SessionPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	var metrics []plugin.MetricType

	fileout := config["file"].(ctypes.ConfigValueStr).Value

	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}

	file, err := os.OpenFile(fileout, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, m := range metrics {
		var keys []string
		for key, value := range(m.Tags()) {
			keys = append(keys, key + "=" + value)
		}


		w.WriteString(fmt.Sprintf("%s\t%s\n", "/" + strings.Join(m.Namespace().Strings(), "/"), strings.Join(keys, ",")))
	}
	w.Flush()

	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func (f *SessionPublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("file", true)
	handleErr(err)
	r1.Description = "Absolute path to the output file for publishing"

	config.Add(r1)
	cp.Add([]string{""}, config)
	return cp, nil
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
