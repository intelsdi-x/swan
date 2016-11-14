package main

import (
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-caffe-inference/caffe"
)

func main() {
	meta := caffe.Meta()
	meta.RPCType = plugin.JSONRPC

	plugin.Start(meta, caffe.NewCaffeInference(time.Now()), os.Args[1])
}
