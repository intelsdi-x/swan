package caffe

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/logger"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/pkg/errors"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME          = "caffeinference"
	VERSION       = 1
	TYPE          = plugin.CollectorPluginType
	UNIT          = "images"
	DESCRIPTION   = "Images classified by caffe"
	imagesInBatch = 10000
)

var (
	namespace = []string{"intel", "swan", "caffe", "inference"}
)

type CaffeInferenceCollector struct {
}

// GetMetricTypes implements plugin.PluginCollector interface.
// Single metric only: /intel/swan/caffe/interference/img which holds number of processed images.
func (CaffeInferenceCollector) GetMetricTypes(configType plugin.ConfigType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType

	namespace := core.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	namespace = namespace.AddStaticElement("img")
	metrics = append(metrics, plugin.MetricType{Namespace_: namespace, Unit_: UNIT, Version_: VERSION})

	return metrics, nil
}

// CollectMetrics implements plugin.PluginCollector interface.
func (CaffeInferenceCollector) CollectMetrics(metricTypes []plugin.MetricType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType

	if len(metricTypes) > 1 {
		msg := "Too much metrics requested. Caffe inference collector gathers single metric."
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

	// [...]string{"intel", "swan", "caffe", "interfere", "hostname", "img"}
	const namespaceHostnameIndex = 4
	const swanNamespacePrefix = 5

	metricType := metricTypes[0]

	metric := plugin.MetricType{Namespace_: metricType.Namespace_,
		Unit_:        metricType.Unit_,
		Version_:     metricType.Version_,
		Description_: DESCRIPTION}
	metric.Namespace_[namespaceHostnameIndex].Value = hostname
	metric.Timestamp_ = time.Now()

	//Parsing caffe output succeeded so images holds value of processed images
	metric.Data_ = images

	metrics = append(metrics, metric)
	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (CaffeInferenceCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
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
func Meta() *plugin.PluginMeta {
	meta := plugin.NewPluginMeta(
		NAME,
		VERSION,
		TYPE,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1*time.Second),
	)
	meta.RPCType = plugin.JSONRPC
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
	scanner.Split(bufio.ScanLines)

	re := regexp.MustCompile("Batch ([0-9]+)")
	if re == nil {
		err = errors.New("Internal error in caffe inference collector")
		return 0, err
	}

	foundBatch := false
	result := uint64(0)
	for scanner.Scan() {
		regexpResult := re.FindAllStringSubmatch(scanner.Text(), -1)
		if regexpResult != nil {
			batchNum, err := strconv.ParseUint(regexpResult[0][1], 10, 64)
			if err == nil {
				result = batchNum * imagesInBatch
				foundBatch = true
			} else {
				err = errors.New("Invalid batch number in the output log")
			}
		}
	}

	if foundBatch != true {
		err = errors.New("Did not find batch number in the output log")
	}
	return result, err
}
