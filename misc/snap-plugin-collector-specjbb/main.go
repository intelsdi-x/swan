package main

import (
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-specjbb/specjbb"
)

func main() {
	meta := specjbb.Meta()
	plugin.Start(meta, specjbb.NewSpecjbb(time.Now()), os.Args[1])
}
