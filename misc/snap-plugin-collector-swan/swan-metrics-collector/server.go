package swanMetricsCollector

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
Error codes:
- 401 - Malformed data
- 402 - Received data isn't JSON
*/

// TODO squall0: Port could be set via environment variable
const (
	collectorPort = 5940
)

func StartServer() {
	http.HandleFunc("/", getRequestContent)
	http.ListenAndServe(fmt.Sprintf(":%d", collectorPort), nil)
}

func getRequestContent(w http.ResponseWriter, r *http.Request) {
	metricsRaw, err := ioutil.ReadAll(r.Body)
	fmt.Printf("DATA: %s\n", metricsRaw)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	if err := updateMetrics(metricsRaw); err != nil {
		w.WriteHeader(402)
		fmt.Printf("ERR %s\n", err)
		return
	}
}
