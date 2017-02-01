package caffe

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/pkg/errors"
)

// Constants representing plugin name, version, type and unit of measurement used.
const (
	NAME        = "caffe-inference"
	VERSION     = 1
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
func (InferenceCollector) GetMetricTypes(configType plugin.Config) ([]plugin.Metric, error) {
	var metrics []plugin.Metric

	namespace := plugin.NewNamespace(namespace...)
	namespace = namespace.AddDynamicElement("hostname", "Name of the host that reports the metric")
	namespace = namespace.AddStaticElement(METRICNAME)
	metrics = append(metrics, plugin.Metric{Namespace: namespace, Unit: UNIT, Version: VERSION})

	return metrics, nil
}

// CollectMetrics implements plugin.PluginCollector interface.
func (InferenceCollector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	var metrics []plugin.Metric

	for _, requestedMetric := range metricTypes {
		requestedMetricNamespace := strings.Join(append([]string{""}, requestedMetric.Namespace.Strings()...), "/")
		pluginPrefixNamespace := strings.Join(append([]string{""}, namespace...), "/")
		if strings.HasPrefix(requestedMetricNamespace, pluginPrefixNamespace) != true {
			log.Errorf("requested metric %q does not match to the caffe inference collector provided metric (prefix) %q", requestedMetric, pluginPrefixNamespace)
			return metrics, ErrNspace
		}

		sourceFileName, err := requestedMetric.Config.GetString("stdout_file")
		if err != nil {
			log.Errorf("stdout_file missing or wrong in config for namespace: %s, error: %s", requestedMetricNamespace, err.Error())
			return metrics, ErrConf
		}

		batches, err := parseOutputFile(sourceFileName)
		if err != nil {
			log.Errorf("parsing caffe output (%s) for namespace %s failed: %s", sourceFileName, requestedMetricNamespace, err.Error())
			return metrics, ErrParse
		}

		// [...]string{"intel", "swan", "caffe", "inference", "hostname", "img"}
		const namespaceHostnameIndex = 4
		const swanNamespacePrefix = 5

		hostname := ""
		if requestedMetric.Namespace.Element(namespaceHostnameIndex).Value == "*" {
			hostname, err = os.Hostname()
			if err != nil {
				log.Errorf("cannot determine hostname: %s", err.Error())
				return metrics, ErrPlugin
			}
		} else {
			hostname = requestedMetric.Namespace.Element(namespaceHostnameIndex).Value
		}

		metric := plugin.Metric{Namespace: requestedMetric.Namespace,
			Unit:        requestedMetric.Unit,
			Version:     requestedMetric.Version,
			Description: DESCRIPTION}
		metric.Namespace[namespaceHostnameIndex].Value = hostname
		metric.Timestamp = time.Now()

		//Parsing caffe output succeeded so images holds value of processed images
		metric.Data = batches

		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// GetConfigPolicy implements plugin.PluginCollector interface.
func (InferenceCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.ConfigPolicy{}
	err := policy.AddNewStringRule([]string{}, "stdout_file", true)
	if err != nil {
		log.Errorf("cannot create new string rule: %s", err.Error())
		return policy, ErrPlugin
	}

	return policy, nil
}

// Meta returns plugin metadata.
//func Meta() *plugin.PluginMeta {
//	meta := plugin.NewPluginMeta(
//		NAME,
//		VERSION,
//		TYPE,
//		[]string{plugin.SnapGOBContentType},
//		[]string{plugin.SnapGOBContentType},
//		plugin.Unsecure(true),
//		plugin.RoutingStrategy(plugin.DefaultRouting),
//		plugin.CacheTTL(1*time.Second),
//	)
//
//	return meta
//}

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
		log.Errorf("cannot read file %s at %v: %s", path, stat.Size()-int64(len(buf)), err.Error())
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
