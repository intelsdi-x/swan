package caffe

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/pkg/errors"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME        = "caffeinference"
	VERSION     = 1
	TYPE        = plugin.CollectorPluginType
	METRICNAME  = "batches"
	UNIT        = "batches"
	DESCRIPTION = "Images classified by caffe (in batches)"
)

var (
	namespace = []string{"intel", "swan", "caffe", "inference"}
	// ErrNspace means that invalid namespace was passed to the plugin
	ErrNspace = errors.New("invalid namespace")
	// ErrConf means that stdout_file is missing
	ErrConf = errors.New("invalid config")
	// ErrParse means that parsing stdout_file failed
	ErrParse = errors.New("parse stdout_file")
	// ErrPlugin means that plugin related error occurred
	ErrPlugin = errors.New("plugin internal error")
)

// InferenceCollector implements snap Plugin interface.
type InferenceCollector struct {
}

// GetMetricTypes implements plugin.PluginCollector interface.
// Single metric only: /intel/swan/caffe/inference/img which holds number of processed images.
func (InferenceCollector) GetMetricTypes(configType plugin.ConfigType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType

	namespace := core.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	namespace = namespace.AddStaticElement(METRICNAME)
	metrics = append(metrics, plugin.MetricType{Namespace_: namespace, Unit_: UNIT, Version_: VERSION})

	return metrics, nil
}

// CollectMetrics implements plugin.PluginCollector interface.
func (InferenceCollector) CollectMetrics(metricTypes []plugin.MetricType) ([]plugin.MetricType, error) {
	var metrics []plugin.MetricType

	for _, requestedMetric := range metricTypes {
		requestedMetricNamespace := requestedMetric.Namespace().String()
		pluginPrefixNamespace := strings.Join(append([]string{""}, namespace...), "/")
		if strings.HasPrefix(requestedMetricNamespace, pluginPrefixNamespace) != true {
			log.Errorf("requested metric %q does not match to the caffe inference collector provided metric (prefix) %q", requestedMetric, pluginPrefixNamespace)
			return metrics, ErrNspace
		}

		sourceFilePath, err := config.GetConfigItem(requestedMetric, "stdout_file")
		if err != nil {
			log.Errorf("stdout_file missing or wrong in config for namespace: %s, error: %s", requestedMetricNamespace, err.Error())
			return metrics, ErrConf
		}

		sourceFileName, ok := sourceFilePath.(string)
		if !ok {
			log.Errorf("stdout_file name invalid for namespace %s", requestedMetricNamespace)
			return metrics, ErrConf
		}

		batches, err := parseOutputFile(sourceFileName)
		if err != nil {
			log.Errorf("parsing caffe output (%s) for namespace %s failed: %s", sourceFilePath.(string), requestedMetricNamespace, err.Error())
			return metrics, ErrParse
		}

		// [...]string{"intel", "swan", "caffe", "inference", "hostname", "img"}
		const namespaceHostnameIndex = 4
		const swanNamespacePrefix = 5

		hostname := ""
		if requestedMetric.Namespace_[namespaceHostnameIndex].Value == "*" {
			hostname, err = os.Hostname()
			if err != nil {
				log.Errorf("cannot determine hostname: %s", err.Error())
				return metrics, ErrPlugin
			}
		} else {
			hostname = requestedMetric.Namespace_[namespaceHostnameIndex].Value
		}

		metric := plugin.MetricType{Namespace_: requestedMetric.Namespace_,
			Unit_:        requestedMetric.Unit_,
			Version_:     requestedMetric.Version_,
			Description_: DESCRIPTION}
		metric.Namespace_[namespaceHostnameIndex].Value = hostname
		metric.Timestamp_ = time.Now()

		//Parsing caffe output succeeded so images holds value of processed images
		metric.Data_ = batches

		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (InferenceCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	policy := cpolicy.New()
	stdoutFile, err := cpolicy.NewStringRule("stdout_file", true)
	if err != nil {
		log.Errorf("cannot create new string rule: %s", err.Error())
		return policy, ErrPlugin
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
	stat, err := os.Stat(path)
	if err != nil {
		log.Errorf("cannot stat file %s: %s", path, err.Error())
		return 0, ErrParse
	}

	file, err := os.Open(path)
	if err != nil {
		log.Errorf("cannot open file %s: %s", path, err.Error())
		return 0, ErrParse
	}
	defer file.Close()

	// In correctly finished log buffer roughly 269 characters is enough
	// to get last Batch XXXX occurence.
	// If caffe was killed the last occurence will be even closer to the
	// EOF. See example output files.
	buf := make([]byte, 4096)

	readat := int64(0)
	if stat.Size() > int64(len(buf)) {
		readat = stat.Size() - int64(len(buf))
	}
	n, err := file.ReadAt(buf, readat)
	if err != nil {
		log.Errorf("cannot cannot read file %s at %v: %s", path, stat.Size()-int64(len(buf)), err.Error())
		return 0, ErrParse
	}

	buf2 := buf[:n]
	scanner := bufio.NewScanner(strings.NewReader(string(buf2)))
	scanner.Split(bufio.ScanLines)

	re := regexp.MustCompile("Batch ([0-9]+)")
	if re == nil {
		log.Errorf("failed to parse: %s, plugin internal error", path)
		return 0, ErrParse
	}

	foundBatch := false
	result := uint64(0)
	for scanner.Scan() {
		regexpResult := re.FindAllStringSubmatch(scanner.Text(), -1)
		if regexpResult != nil {
			batchNum, err := strconv.ParseUint(regexpResult[0][1], 10, 64)
			if err == nil {
				result = batchNum
				foundBatch = true
			}
		}
	}

	if foundBatch != true {
		log.Errorf("failed to parse: %s, did not find valid batch number", path)
		err = ErrParse
	}
	return result, err
}
