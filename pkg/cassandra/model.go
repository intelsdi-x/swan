package cassandra

import "time"

// Metrics is a struct with all values stored in keyspace snap.metrics in cassandra.
type Metrics struct {
	namespace string
	version   int
	host      string
	time      time.Time
	boolval   bool
	doubleval float64
	labels    []string
	strval    string
	tags      map[string]string
	valtype   string
}

// NewMetrics returns a new Metrics struct.
func NewMetrics(namespace string, version int, host string, time time.Time, boolval bool, doubleval float64,
	labels []string, strval string, tags map[string]string, valtype string) *Metrics {
	return &Metrics{
		namespace,
		version,
		host,
		time,
		boolval,
		doubleval,
		labels,
		strval,
		tags,
		valtype,
	}
}

// Namespace returns namespace from given Metrics struct.
func (metric *Metrics) Namespace() string {
	return metric.namespace
}

// Version returns version of plugin from given Metrics struct.
func (metric *Metrics) Version() int {
	return metric.version
}

// Host returns host from given Metrics struct.
func (metric *Metrics) Host() string {
	return metric.host
}

// Time returns time from given Metrics struct.
func (metric *Metrics) Time() time.Time {
	return metric.time
}

// Boolval returns boolval from given Metrics struct.
func (metric *Metrics) Boolval() bool {
	return metric.boolval
}

// Doubleval returns doubleval from given Metrics struct.
func (metric *Metrics) Doubleval() float64 {
	return metric.doubleval
}

// Labels returns labels from given Metrics struct.
func (metric *Metrics) Labels() []string {
	return metric.labels
}

// Strval returns strval from given Metrics struct.
func (metric *Metrics) Strval() string {
	return metric.strval
}

// Tags returns tags from given Metrics struct.
func (metric *Metrics) Tags() map[string]string {
	return metric.tags
}

// Valtype returns valtype from given Metrics struct.
func (metric *Metrics) Valtype() string {
	return metric.valtype
}
