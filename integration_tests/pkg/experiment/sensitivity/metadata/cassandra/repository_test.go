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
		phaseTable := cassandra.NewPhaseTable(keySpace)
		measurementTable := cassandra.NewMeasurementTable(keySpace)
		cassandra := cassandra.NewCassandra(cassandra.NewExperimentTable(keySpace), phaseTable, measurementTable)
		Convey("I should get an Uploader instance", func() {
			So(cassandra, ShouldNotBeNil)
			Convey("When I pass SwanMetrics instance", func() {
				sM := createValidSwanMetrics()
				err = cassandra.Save(sM)
				So(err, ShouldBeNil)
				Convey("I should get no error and see experiment metadata saved", func() {
					experiment, err := cassandra.Fetch("experiment")
					So(err, ShouldBeNil)

					So(experiment.ID, ShouldEqual, "experiment")
					So(experiment.LCName, ShouldEqual, "Latency critical task")
					So(experiment.LGName, ShouldEqual, "Load generator task")
					So(experiment.LoadPointsNumber, ShouldEqual, 509)
					So(experiment.RepetitionsNumber, ShouldEqual, 905)

					oneSecond, _ := time.ParseDuration("1s")
					So(experiment.LoadDuration, ShouldEqual, oneSecond)
					twoSeconds, _ := time.ParseDuration("2s")
					So(experiment.TuningDuration, ShouldEqual, twoSeconds)

					Convey("I should get no error and see phases metadata saved", func() {
						So(experiment.Phases, ShouldHaveLength, 1)
						So(experiment.Phases[0].ID, ShouldEqual, "phase")

						So(experiment.Phases[0].Aggressors, ShouldHaveLength, 1)
						So(experiment.Phases[0].Aggressors[0].Isolation, ShouldEqual, "aggressor isolation")
						So(experiment.Phases[0].Aggressors[0].Name, ShouldEqual, "an aggressor")
						So(experiment.Phases[0].Aggressors[0].Parameters, ShouldEqual, "aggressor parameters")
						So(experiment.Phases[0].LCIsolation, ShouldEqual, "Latency critical isolation")
						So(experiment.Phases[0].LCParameters, ShouldEqual, "Latency critical parameters")

						Convey("I should get no error and see measurement metadata saved", func() {
							So(experiment.Phases[0].Measurements, ShouldHaveLength, 1)
							So(experiment.Phases[0].Measurements[0].Load, ShouldEqual, 65)
							So(experiment.Phases[0].Measurements[0].LoadPointQPS, ShouldEqual, 666)
							So(experiment.Phases[0].Measurements[0].LGParameters, ShouldEqual, "Load generator parameters")
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
					brakeDatabase(config, "DROP TABLE", phaseTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Phases metadata fetch failed")
					So(metadata.ID, ShouldEqual, "")
				})
				Convey("I should get en error when phase metadata query returns no results", func() {
					brakeDatabase(config, "TRUNCATE", phaseTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Phases metadata fetch returned no results")
					So(metadata.ID, ShouldEqual, "")

				})
				Convey("I should get an error when measurement metadata query fails", func() {
					brakeDatabase(config, "DROP TABLE", measurementTable.Name())
					metadata, err := cassandra.Fetch("experiment")
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "Measurements metadata fetch failed")
					So(metadata.ID, ShouldEqual, "")
				})
				Convey("I should get en error when measurement metadata query returns no results", func() {
					brakeDatabase(config, "TRUNCATE", measurementTable.Name())
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
			query.Release()
			session.Close()
			So(err, ShouldBeNil)
		})
	})
}

func createValidSwanMetrics() metadata.Experiment {
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

func brakeDatabase(config cassandra.Config, operation, tableName string) {
	session, err := createGocqlSession(config)
	So(err, ShouldBeNil)
	query := session.Query(fmt.Sprintf("%s %s.%s", operation, config.KeySpace, tableName))
	err = query.Exec()
	query.Release()
	session.Close()
	So(err, ShouldBeNil)
}
