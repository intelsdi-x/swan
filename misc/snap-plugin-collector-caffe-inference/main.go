package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-caffe-inference/caffe"
)

func main() {
	meta := caffe.Meta()
	meta.RPCType = plugin.JSONRPC

	plugin.Start(meta, caffe.CaffeInferenceCollector{}, os.Args[1])
}
