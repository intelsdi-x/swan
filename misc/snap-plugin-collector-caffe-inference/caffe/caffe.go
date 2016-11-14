package caffe

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/logger"
	snapPlugin "github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/pkg/errors"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME          = "caffeinference"
	VERSION       = 1
	TYPE          = snapPlugin.CollectorPluginType
	UNIT          = "ns"
	imagesInBatch = 10000
)

var (
	namespace = []string{"intel", "swan", "caffeinference"}
)

type plugin struct {
	now time.Time
}

// NewMutilate creates new mutilate collector.
func NewCaffeInference(now time.Time) snapPlugin.CollectorPlugin {
	return &plugin{now}
}

// GetMetricTypes implements plugin.PluginCollector interface.
// Single metric only: /intel/swan/caffe/interference/img which holds number of processed images.
func (mutilate *plugin) GetMetricTypes(configType snapPlugin.ConfigType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	namespace := core.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	namespace = namespace.AddStaticElement("img")
	metrics = append(metrics, snapPlugin.MetricType{Namespace_: namespace, Unit_: UNIT, Version_: VERSION})

	return metrics, nil
}

// CollectMetrics implements plugin.PluginCollector interface.
func (mutilate *plugin) CollectMetrics(metricTypes []snapPlugin.MetricType) ([]snapPlugin.MetricType, error) {
	var metrics []snapPlugin.MetricType

	if len(metricTypes) > 1 {
		msg := fmt.Sprintf("Too much metrics requested. Caffe inference collector gathers single metric.")
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	sourceFilePath, err := config.GetConfigItem(metricTypes[0], "stdout_file")
	if err != nil {
		msg := fmt.Sprintf("No file path set - no metrics are going to be collected: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	images, err := parseOutputFile(sourceFilePath.(string))
	if err != nil {
		msg := fmt.Sprintf("Parsing caffe output failed: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	hostname, err := os.Hostname()
	if err != nil {
		msg := fmt.Sprintf("Cannot determine hostname: %s", err.Error())
		logger.LogError(msg)
		return metrics, errors.New(msg)
	}

	// [...]string{"intel", "swan", "caffe-interfere", "hostname"}
	const namespaceHostnameIndex = 3
	const swanNamespacePrefix = 4

	metricType := metricTypes[0]

	metric := snapPlugin.MetricType{Namespace_: metricType.Namespace_, Unit_: metricType.Unit_, Version_: metricType.Version_}
	metric.Namespace_[namespaceHostnameIndex].Value = hostname
	metric.Timestamp_ = mutilate.now

	//Parsing caffe output succeeded so images holds value of processed images
	metric.Data_ = images

	metrics = append(metrics, metric)
	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (mutilate *plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	policy := cpolicy.New()
	stdoutFile, err := cpolicy.NewStringRule("stdout_file", true)
	if err != nil {
		return policy, errors.Wrap(err, "cannot create new string rule")
	}
	policyNode := cpolicy.NewPolicyNode()
	policyNode.Add(stdoutFile)
	policy.Add(namespace, policyNode)

	return policy, nil
}

// Meta returns plugin metadata.
func Meta() *snapPlugin.PluginMeta {
	meta := snapPlugin.NewPluginMeta(
		NAME,
		VERSION,
		TYPE,
		[]string{snapPlugin.SnapGOBContentType},
		[]string{snapPlugin.SnapGOBContentType},
		snapPlugin.Unsecure(true),
		snapPlugin.RoutingStrategy(snapPlugin.DefaultRouting),
		snapPlugin.CacheTTL(1*time.Second),
	)
	meta.RPCType = snapPlugin.JSONRPC
	return meta
}

// Longest valid output will look like following:
// I1109 13:24:05.241741  2329 caffe.cpp:275] Batch 99, loss = 0.75406
// I1109 13:24:05.241747  2329 caffe.cpp:280] Loss: 0.758892
// I1109 13:24:05.241760  2329 caffe.cpp:292] accuracy = 0.7515
// I1109 13:24:05.241771  2329 caffe.cpp:292] loss = 0.758892 (* 1 = 0.758892 loss)
// Therefore looking back 269 characters and searching for word 'Batch' is
// sufficient.
func parseOutputFile(path string) (uint64, error) {

	fmt.Printf("Opening file %s\n", path)

	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// In correctly finished log buffer of this size guarantees occurence
	// Of the word 'Batch'. If caffe was killed then there will be many
	// occurences.
	buf := make([]byte, 269)

	n, err := file.ReadAt(buf, stat.Size()-int64(len(buf)))
	if err != nil {
		return 0, err
	}

	buf2 := buf[:n]
	scanner := bufio.NewScanner(strings.NewReader(string(buf2)))
	scanner.Split(bufio.ScanWords)

	// Global flag if 'Batch' word was found
	foundBatch := 0
	// Local flag. After every 'Batch' number is expected
	batch := 0
	// Keeps track of last known 'Batch' value
	lastBatch := ""
	for scanner.Scan() {
		if scanner.Text() == "Batch" {
			foundBatch++
			batch = 1
			continue
		}
		if foundBatch > 0 && batch == 1 {
			// Get rid of the comma in 'number,' token.
			lastBatch = strings.Replace(scanner.Text(), ",", "", 1)
			foundBatch++
		}
		batch = 0
	}
	result := uint64(0)
	if foundBatch > 1 {
		batchNum, err := strconv.ParseUint(lastBatch, 10, 64)
		if err == nil {
			result = batchNum * imagesInBatch
		} else {
			err = errors.New("Invalid batch number in the output log")
		}
	} else {
		err = errors.New("Did not find batch number in the output log")
	}
	return result, err
}
