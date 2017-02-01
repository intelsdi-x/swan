package main

import (
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-caffe-inference/caffe"
)

func main() {
	plugin.StartCollector(caffe.InferenceCollector{}, caffe.NAME, caffe.VERSION)
}
