package sessionPublisher

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/pkg/errors"
)

const (
	// NAME is name of the plugin used to register it in Snap.
	NAME = "session-test"
	// VERSION represents version of the plugin.
	VERSION = 1
)

// The SessionPublisher is a test publisher hosted in swan which enables
// the session test to verify that tags have indeed been added to the metrics.
type SessionPublisher struct{}

// Publish is an implementation needed for the Publisher interface and here,
// stores metrics by namespace and tags to a file, defined in the plugin configuration.
func (f SessionPublisher) Publish(inputMetrics []plugin.Metric, config plugin.Config) error {
	var metrics []plugin.Metric

	fileout, err := config.GetString("file")
	if err != nil {
		errors.Wrap(err, "Unable to retrive file from configuration")
	}

	file, err := os.OpenFile(fileout, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, m := range metrics {
		var tags []string
		for key, value := range m.Tags {
			tags = append(tags, key+"="+value)
		}

		// Make row: Namespace\t Tags\t Values\n.
		w.WriteString(
			fmt.Sprintf(
				"%s\t%s\t%v\n",
				"/"+strings.Join(m.Namespace.Strings(), "/"),
				strings.Join(tags, ","),
				m.Data,
			))
	}
	w.Flush()

	return nil
}

// GetConfigPolicy is an implementation needed for the Publisher interface and here,
// returns configuration requiring 'swan_experiment' and 'swan_phase' to be set.
func (f SessionPublisher) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	cp := plugin.ConfigPolicy{}
	err := cp.AddNewStringRule([]string{""}, "file", true)
	if err != nil {
		panic(err)
	}
	return cp, nil
}

// Meta returns a plugin meta data.
//func Meta() *plugin.PluginMeta {
//	return plugin.NewPluginMeta(NAME, VERSION, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
//}
