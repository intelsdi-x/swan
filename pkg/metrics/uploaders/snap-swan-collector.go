package uploaders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/intelsdi-x/swan/pkg/metrics"
)

const (
	// TODO squall0: Add external configurator to this parameters
	// TODO squall0: Host value is a Snap host value
	// TODO squall0: Port could be passed as a Environmental Variable
	defaultCollectorHost = "127.0.0.1"
	defaultCollectorPort = 5940
)

func SendMetrics(s SwanMetrics) error {

	collectorUrl := fmt.Sprintf("http://%s:%d", defaultCollectorHost, defaultCollectorPort)

	buffer, err := json.Marshal(s)
	if err != nil {
		return err
	}

	res, err := http.Post(collectorUrl, "application/javascript", bytes.NewReader(buffer))
	defer res.Body.Close()

	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Wrong exit code: %d. %s", res.StatusCode, res.Status)
	}

	return nil
}
