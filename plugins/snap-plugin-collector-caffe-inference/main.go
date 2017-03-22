package main

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-collector-caffe-inference/caffe"
)

func main() {
	plugin.StartCollector(caffe.InferenceCollector{}, caffe.NAME, caffe.VERSION, plugin.CacheTTL(1*time.Second))
}
