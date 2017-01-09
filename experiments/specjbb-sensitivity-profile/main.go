package main

import (
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

func main() {
	err := common.RunExperiment()
	errutil.Check(err)
}
