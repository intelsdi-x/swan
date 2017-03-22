package main

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-collector-specjbb/specjbb"
)

func main() {
	plugin.StartCollector(specjbb.NewSpecjbb(time.Now()), specjbb.NAME, specjbb.VERSION, plugin.CacheTTL(1*time.Second))
}
