// Copyright (c) 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/metadata"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"os"
	"time"
	"github.com/intelsdi-x/swan/pkg/conf"
	"google.golang.org/grpc"
	kricoapi "github.com/intelsdi-x/swan/experiments/krico/api"
	"context"
	"github.com/intelsdi-x/swan/experiments/krico/workloads"
	"strings"
)

var (
	appName = os.Args[0]
	kricoApiAddress = conf.NewStringFlag("krico_api_address", "Ip address of KRICO API service.", "localhost:5000")
)

func main() {

	// Preparing application - setting name, help, parsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	experimentID := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, experimentID)

	// Connect to metadata database.
	metaData, err := metadata.NewDefault(experimentID)
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(metaData, experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Initialize workloads.
	workload.Initialize(experimentID)

	// Run workloads.
	workload.CachingWorkload()

	// Prepare metadata.
	records := map[string]string{
		experiment.ExperimentKey: 	experimentID,
		"commands_arguments": 		strings.Join(os.Args, ","),
		"experiment_name":    		appName,
	}

	// Save metadata.
	err = metaData.RecordMap(records, metadata.TypeEmpty)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")

	// Connect to KRICO via gRPC to update metadata.
	conn, err := grpc.Dial(kricoApiAddress.Value(), grpc.WithInsecure())
	errutil.CheckWithContext(err, "Cannot connect to KRICO gRPC server")

	api := kricoapi.NewApiClient(conn)

	_, err = api.LoadSwanExperiment(context.Background(), &kricoapi.LoadSwanExperimentRequest{ExperimentId: experimentID})
	errutil.CheckWithContext(err, "Couldn't connect to KRICO api via grpcs")

	_, err = api.RefreshClassifier(context.Background(), &kricoapi.RefreshClassifierRequest{})
	_, err = api.RefreshPredictor(context.Background(), &kricoapi.RefreshPredictorRequest{})

	defer conn.Close()
}
