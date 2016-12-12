package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-caffe-inference/caffe"
)

func main() {
	meta := caffe.Meta()

	plugin.Start(meta, caffe.InferenceCollector{}, os.Args[1])
}
