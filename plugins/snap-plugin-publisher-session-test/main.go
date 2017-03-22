package main

import (
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-publisher-session-test/session-publisher"
)

func main() {
	plugin.StartPublisher(sessionPublisher.SessionPublisher{}, sessionPublisher.NAME, sessionPublisher.VERSION)
}
