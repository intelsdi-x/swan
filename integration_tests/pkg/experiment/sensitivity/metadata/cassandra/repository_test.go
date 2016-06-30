package cassandra

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata/cassandra"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraUploading(t *testing.T) {
	Convey("When I connect to Cassandra instance with incorrect configuration", t, func() {
		config := cassandra.Config{}
		cassandra, err := cassandra.NewKeySpace(config)
		Convey("I should get an error and no Uploader instance", func() {
			So(err, ShouldNotBeNil)
			So(cassandra, ShouldBeNil)
		})
	})

	Convey("When I connect to Cassandra instance with correct configuration", t, func() {
		config := cassandra.Config{}
		config.Host = []string{"127.0.0.1"}
		config.KeySpace = "testing_keyspace"

		keySpace, err := cassandra.NewKeySpace(config)
		So(err, ShouldBeNil)
		experimentTable, err := cassandra.NewExperimentTable(keySpace)
		So(err, ShouldBeNil)
		phaseTable, err := cassandra.NewPhaseTable(keySpace)
		So(err, ShouldBeNil)
		measurementTable, err := cassandra.NewMeasurementTable(keySpace)
		So(err, ShouldBeNil)
		cassandra := cassandra.NewCassandra(experimentTable, phaseTable, measurementTable)
		Convey("I should get an Uploader instance", func() {
			So(cassandra, ShouldNotBeNil)
			Convey("When I pass SwanMetrics instance", func() {
				metadata := createMetadata()
				err = cassandra.Save(metadata)
				So(err, ShouldBeNil)
				Convey("I should get no error and see experiment metadata saved", func() {
					experiment, err := cassandra.Fetch("experiment")
					So(err, ShouldBeNil)

					So(experiment.ID, ShouldEqual, metadata.ID)
					So(experiment.LCName, ShouldEqual, metadata.LCName)
					So(experiment.LGName, ShouldEqual, metadata.LGName)
					So(experiment.LoadPointsNumber, ShouldEqual, metadata.LoadPointsNumber)
					So(experiment.RepetitionsNumber, ShouldEqual, metadata.RepetitionsNumber)

					So(experiment.LoadDuration, ShouldEqual, metadata.LoadDuration)
					So(experiment.TuningDuration, ShouldEqual, metadata.TuningDuration)

					Convey("I should get no error and see phases metadata saved", func() {
						So(experiment.Phases, ShouldHaveLength, 1)
						So(experiment.Phases[0].ID, ShouldEqual, metadata.Phases[0].ID)

						So(experiment.Phases[0].Aggressors, ShouldHaveLength, 1)
						So(experiment.Phases[0].Aggressors[0].Isolation, ShouldEqual, metadata.Phases[0].Aggressors[0].Isolation)
						So(experiment.Phases[0].Aggressors[0].Name, ShouldEqual, metadata.Phases[0].Aggressors[0].Name)
						So(experiment.Phases[0].Aggressors[0].Parameters, ShouldEqual, metadata.Phases[0].Aggressors[0].Parameters)
						So(experiment.Phases[0].LCIsolation, ShouldEqual, metadata.Phases[0].LCIsolation)
						So(experiment.Phases[0].LCParameters, ShouldEqual, metadata.Phases[0].LCParameters)

						Convey("I should get no error and see measurement metadata saved", func() {
							So(experiment.Phases[0].Measurements, ShouldHaveLength, 1)
							So(experiment.Phases[0].Measurements[0].Load, ShouldEqual, metadata.Phases[0].Measurements[0].Load)
							So(experiment.Phases[0].Measurements[0].LoadPointQPS, ShouldEqual, metadata.Phases[0].Measurements[0].LoadPointQPS)
							So(experiment.Phases[0].Measurements[0].LGParameters, ShouldEqual, metadata.Phases[0].Measurements[0].LGParameters)
						})
					})
				})
				Convey("I should get an error when I try to get metadata for non existing experiment", func() {
					metadata, err := cassandra.Fetch("non existing experiment")
					So(err.Error(), ShouldStartWith, "Experiment metadata fetch failed")
					So(err, ShouldNotBeNil)
					So(metadata.ID, ShouldEqual, "")
				})
				Convey("I should get an error when phase metadata query fails", func() {
					corruptDatabase(config, "DROP TABLE", phaseTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Phases metadata fetch failed")
					So(metadata.ID, ShouldEqual, "")
				})
				Convey("I should get en error when phase metadata query returns no results", func() {
					corruptDatabase(config, "TRUNCATE", phaseTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Phases metadata fetch returned no results")
					So(metadata.ID, ShouldEqual, "")

				})
				Convey("I should get an error when measurement metadata query fails", func() {
					corruptDatabase(config, "DROP TABLE", measurementTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Measurements metadata fetch failed")
					So(metadata.ID, ShouldEqual, "")
				})
				Convey("I should get en error when measurement metadata query returns no results", func() {
					corruptDatabase(config, "TRUNCATE", measurementTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Measurements metadata fetch returned no results")
					So(metadata.ID, ShouldEqual, "")

				})

			})
		})

		Reset(func() {
			time.Sleep(1 * time.Second)
			session, err := createGocqlSession(config)
			So(err, ShouldBeNil)
			query := session.Query(fmt.Sprintf("DROP KEYSPACE %s", config.KeySpace))
			err = query.Exec()
			So(err, ShouldBeNil)
			query.Release()
			session.Close()
		})
	})
}

func createMetadata() metadata.Experiment {
	oneSecond, _ := time.ParseDuration("1s")
	twoSeconds, _ := time.ParseDuration("2s")
	meta := metadata.Experiment{
		BaseExperiment: metadata.BaseExperiment{
			RepetitionsNumber: 905,
			LoadPointsNumber:  509,
			LCName:            "Latency critical task",
			LGName:            "Load generator task",
			ID:                "experiment",
			LoadDuration:      oneSecond,
			TuningDuration:    twoSeconds,
		},
	}

	meta.AddPhase(metadata.Phase{
		BasePhase: metadata.BasePhase{
			ID:           "phase",
			LCParameters: "Latency critical parameters",
			LCIsolation:  "Latency critical isolation",
		},
		Aggressors: []metadata.Aggressor{metadata.Aggressor{
			Name:       "an aggressor",
			Parameters: "aggressor parameters",
			Isolation:  "aggressor isolation",
		}},
	})

	meta.Phases[0].AddMeasurement(metadata.Measurement{
		Load:         65,
		LoadPointQPS: 666,
		LGParameters: "Load generator parameters",
	})

	return meta
}

func createGocqlSession(config cassandra.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Host...)
	cluster.ProtoVersion = 4
	cluster.Keyspace = config.KeySpace
	cluster.Timeout = 100 * time.Second
	return cluster.CreateSession()

}

func corruptDatabase(config cassandra.Config, operation, tableName string) {
	session, err := createGocqlSession(config)
	So(err, ShouldBeNil)
	query := session.Query(fmt.Sprintf("%s %s.%s", operation, config.KeySpace, tableName))
	err = query.Exec()
	query.Release()
	session.Close()
	So(err, ShouldBeNil)
}
