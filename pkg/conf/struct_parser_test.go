package conf

import (
	"fmt"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestNameFromFieldName(t *testing.T) {
	testData := map[string]string{
		"StringArg":  "string_arg",
		"string_arg": "string_arg",
		"STRINGARG":  "stringarg",
		"STRING_ARG": "string_arg",
		"StringARG":  "string_arg",
		"String1Arg": "string_1_arg",
		"StringArg1": "string_arg_1",
	}

	for fieldName, expectedResult := range testData {
		Convey(fmt.Sprintf("I should get the name = %q from field name = %q", expectedResult, fieldName), t, func() {
			name := nameFromFieldName(fieldName)
			So(name, ShouldEqual, expectedResult)
		})
	}
}

type CorrectTestConfig struct {
	StringArg         string `help:"test string" default:"default_string"`
	StringArg2        string `help:"test string" defaultFromField:"defaultStringArg2"`
	defaultStringArg2 string
	RequiredStringArg string `help:"test required string" required:"true"`
	ExcludedStringArg string

	IntArg         int           `help:"test int" default:"2"`
	DurationArg    time.Duration `help:"test duration" default:"5s"`
	BoolArg        bool          `help:"test bool" default:"true"`
	StringSliceArg []string      `help:"test slice"`
	FileArg        string        `help:"test file" type:"file" defaultFromField:"defaultFileArg"`
	defaultFileArg string
	IPArg          string `help:"test IP" type:"ip" default:"255.255.255.255"`

	// Prefix optional field.
	flagPrefix string
}

func setEnvFromFieldName(fieldName, value string) error {
	flagID := nameFromFieldName(fieldName)
	flag := definedFlags[flagID]
	if flag == nil {
		return errors.Errorf("No flag is defined with id: %s", flagID)
	}

	return os.Setenv(flag.envName(), value)
}

// clearFlags clears the flag which were defined already. It make a clean start for conf tests.
func clearFlags() {
	for _, flag := range definedFlags {
		flag.clear()
	}
	// Modifying a package variable.
	definedFlags = map[string]flagType{}

	app = kingpin.New("test", "No help available")
}

func TestStructTagFlags(t *testing.T) {
	clearFlags()
	Convey("While using Conf flags", t, func() {
		Convey("When a struct exposes fields by using struct tags", func() {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "structTag")
			So(err, ShouldBeNil)

			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			// Process struct.
			config := &CorrectTestConfig{
				defaultStringArg2: "default_string2",
				defaultFileArg:    tmpFile.Name(),
				flagPrefix:        "test",
			}

			Convey("They should have the built-in default values before processing", func() {
				So(config.StringArg, ShouldEqual, "")
				So(config.StringArg2, ShouldEqual, "")
				So(config.RequiredStringArg, ShouldEqual, "")
				So(config.ExcludedStringArg, ShouldEqual, "")
				So(config.IntArg, ShouldEqual, 0)
				So(config.DurationArg, ShouldResemble, 0*time.Millisecond)
				So(config.BoolArg, ShouldEqual, false)
				So(config.StringSliceArg, ShouldResemble, []string(nil))
				So(config.FileArg, ShouldEqual, "")
				So(config.IPArg, ShouldEqual, "")

				err := Process(config)
				So(err, ShouldBeNil)

				Convey("After registration they should have custom default values", func() {
					So(config.StringArg, ShouldEqual, "default_string")
					So(config.StringArg2, ShouldEqual, "default_string2")
					// It is required so no default value.
					So(config.RequiredStringArg, ShouldEqual, "")
					// It should be excluded so NO default value.
					So(config.ExcludedStringArg, ShouldEqual, "")
					So(config.IntArg, ShouldEqual, 2)
					So(config.DurationArg, ShouldResemble, 5*time.Second)
					So(config.BoolArg, ShouldEqual, true)
					So(config.StringSliceArg, ShouldResemble, []string{})
					So(config.FileArg, ShouldEqual, tmpFile.Name())
					So(config.IPArg, ShouldEqual, "255.255.255.255")

					err = ParseEnv()
					// It should be error, since RequiredStringArg is required.
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "required flag --test_required_string_arg not provided")

					// We need to define RequiredStringArg since it is required.
					err = setEnvFromFieldName(config.flagPrefix+"RequiredStringArg", "custom_string")
					So(err, ShouldBeNil)

					// Re-parse.
					err = ParseEnv()
					So(err, ShouldBeNil)

					err = Process(config)
					So(err, ShouldBeNil)

					Convey("Process should reset flags to default values", func() {
						So(config.StringArg, ShouldEqual, "default_string")
						So(config.StringArg2, ShouldEqual, "default_string2")
						So(config.RequiredStringArg, ShouldEqual, "custom_string")
						// It should be excluded so NO custom default value.
						So(config.ExcludedStringArg, ShouldEqual, "")
						So(config.IntArg, ShouldEqual, 2)
						So(config.DurationArg, ShouldResemble, 5*time.Second)
						So(config.BoolArg, ShouldEqual, true)
						So(config.StringSliceArg, ShouldResemble, []string{})
						So(config.FileArg, ShouldEqual, tmpFile.Name())
						So(config.IPArg, ShouldEqual, "255.255.255.255")

						// Define some values.
						err = setEnvFromFieldName(config.flagPrefix+"StringArg", "custom_string")
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"StringArg2", "custom_string2")
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"IntArg", "4324")
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"DurationArg", "10ms")
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"BoolArg", "false")
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"StringSliceArg", "A,B,C,D")
						So(err, ShouldBeNil)

						tmpFile2, err := ioutil.TempFile(os.TempDir(), "structTag")
						So(err, ShouldBeNil)
						defer func() {
							tmpFile2.Close()
							os.Remove(tmpFile2.Name())
						}()
						err = setEnvFromFieldName(config.flagPrefix+"FileArg", tmpFile2.Name())
						So(err, ShouldBeNil)
						err = setEnvFromFieldName(config.flagPrefix+"IPArg", "255.255.255.200")
						So(err, ShouldBeNil)

						// This field should not be exposed.
						err = setEnvFromFieldName(config.flagPrefix+"ExcludedStringArg", "custom_string")
						So(err, ShouldNotBeNil)

						err = ParseEnv()
						So(err, ShouldBeNil)

						err = Process(config)
						So(err, ShouldBeNil)

						Convey("After Parse & Process flags should have custom values", func() {
							So(config.StringArg, ShouldEqual, "custom_string")
							So(config.StringArg2, ShouldEqual, "custom_string2")
							So(config.RequiredStringArg, ShouldEqual, "custom_string")
							// It should be excluded so NO custom default value.
							So(config.ExcludedStringArg, ShouldEqual, "")
							So(config.IntArg, ShouldEqual, 4324)
							So(config.DurationArg, ShouldResemble, 10*time.Millisecond)
							So(config.BoolArg, ShouldEqual, false)
							So(config.StringSliceArg, ShouldResemble, []string{"A", "B", "C", "D"})
							So(config.FileArg, ShouldEqual, tmpFile2.Name())
							So(config.IPArg, ShouldEqual, "255.255.255.200")
						})
					})
				})
			})
		})
	})
}

type TestConfigWithWrongIntDefault struct {
	IntArg int `help:"test int" default:"not an Int"`
}

type TestConfigWithWrongDurationDefault struct {
	DurationArg time.Duration `help:"test duration" default:"not a Duration"`
}

type TestConfigWithWrongBoolDefault struct {
	BoolArg bool `help:"test bool" default:"not a Bool"`
}

type TestConfigWithWrongSliceType struct {
	StringSliceArg []int `help:"test slice"`
}

type TestConfigWithFileDefaultWhichDoesNotExist struct {
	FileArg string `help:"test file" type:"file" default:"/etc/notExistingFile"`
}

type TestConfigWithWrongIPDefault struct {
	IPArg string `help:"test IP" type:"ip" default:"255.255.255.459"`
}

type TestConfigWithUnsupportedType struct {
	FloatArg float64 `help:"this flag should not be supported"`
}

func TestIncorrectStructTags(t *testing.T) {
	Convey("While using Conf flags", t, func() {
		clearFlags()
		Convey("When we specify IntFlag in struct with not parsable default we expect error", func() {
			err := Process(&TestConfigWithWrongIntDefault{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "Wrong default value for Int type flag")
		})

		Convey("When we specify DurationFlag in struct with not parsable default we expect error", func() {
			err := Process(&TestConfigWithWrongDurationDefault{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "Wrong default value for Duration type flag")
		})

		Convey("When we specify BoolFlag in struct with not parsable default we expect error", func() {
			err := Process(&TestConfigWithWrongBoolDefault{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "Wrong default value for Bool type flag")
		})

		Convey("When we specify SliceFlag in struct with not string elements we expect error", func() {
			err := Process(&TestConfigWithWrongSliceType{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "[]int type not supported for a slice flag")
		})

		Convey("When we specify Flag in struct using not supported type like float we expect error", func() {
			err := Process(&TestConfigWithUnsupportedType{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "float64 type not supported for a flag")
		})

		Convey("When we specify FileFlag in struct with default containing not existing file we expect error", func() {
			err := Process(&TestConfigWithFileDefaultWhichDoesNotExist{})
			So(err, ShouldBeNil)

			err = ParseEnv()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEndWith, "Could not parse environment flags: path '/etc/notExistingFile' does not exist")
		})

		Convey("When we specify IPFlag in struct with default containing not parsable IP address we expect error", func() {
			err := Process(&TestConfigWithWrongIPDefault{})
			So(err, ShouldBeNil)

			err = ParseEnv()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEndWith, "is not an IP address")
		})
	})
}
