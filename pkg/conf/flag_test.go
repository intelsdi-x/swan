package conf

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEnvFlag(t *testing.T) {
	Convey("While using Flag struct, it should construct proper swan environment var name", t, func() {
		So(NewStringFlag("test_name", "", "").envName(), ShouldEqual, "SWAN_TEST_NAME")
	})
}

func TestFlags(t *testing.T) {
	Convey("While using Conf flags", t, func() {
		Convey("When some custom String Flag is defined", func() {
			// Register custom flag.
			customFlag := NewStringFlag("custom_string_arg", "help", "default")
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, "default")
			})

			Convey("When we do not define any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "customContent"
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Int Flag is defined", func() {
			// Register custom flag.
			customFlag := NewIntFlag("custom_int_arg", "help", 23424)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, 23424)
			})

			Convey("When we do not define any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := 12
				os.Setenv(customFlag.envName(), fmt.Sprintf("%d", customValue))

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Slice Flag is defined", func() {
			// Register custom flag.
			customFlag := NewSliceFlag("custom_slice_arg", "help")
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldResemble, []string{})
			})

			Convey("When we do not define any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldResemble, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := fmt.Sprintf("A%sB%sC", stringListDelimiter, stringListDelimiter)
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldResemble, []string{"A", "B", "C"})
			})
		})

		Convey("When some custom Bool Flag is defined", func() {
			// Register custom flag.
			customFlag := NewBoolFlag("custom_bool_arg", "help", false)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, false)
			})

			Convey("When we do not define any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := true
				os.Setenv(customFlag.envName(), fmt.Sprintf("%v", customValue))

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Duration Flag is defined", func() {
			// Register custom flag.
			customFlag := NewDurationFlag("custom_duration_arg", "help", 99*time.Millisecond)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, 99*time.Millisecond)
			})

			Convey("When we do not define any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := 1234 * time.Second
				os.Setenv(customFlag.envName(), customValue.String())

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})
	})
}

func TestConfiguration(t *testing.T) {
	Convey("While using flags, we can extract right values for differnt types.", t, func() {

		// Prepare all kinds of flags.
		defaultString := "http://foo-bar"
		stringTestFlag := NewStringFlag("stringTest", "stringDesc", defaultString)
		providedString := "bar-foo"

		defaultInt := 628
		intTestFlag := NewIntFlag("intTest", "intDesc", defaultInt)
		providedInt := "13"

		defaultDuration := 123 * time.Second
		durTestFlag := NewDurationFlag("durationTest", "durDesc", defaultDuration)
		providedDuration := "2h0m0s"

		sliceTestFlag := NewSliceFlag("sliceTest", "sliceDesc")
		providedSlice := "foo1,foo2"

		_, err := app.Parse([]string{
			"--intTest", providedInt,
			"--durationTest", providedDuration,
			"--stringTest", providedString,
			"--sliceTest", providedSlice,
		})
		So(err, ShouldBeNil)

		// External interface (just returns current value by name).
		flagMap := GetFlags()

		// Gather configuration and put into map (for testing purposes).
		// Prepare map with all flags for easier assertions.
		flags := map[string]struct{ Name, Value, Default, Help string }{}
		for _, flag := range getFlagsDefinition() {
			flags[flag.Name] = flag
		}

		// string
		flag, ok := flags[stringTestFlag.name]
		So(ok, ShouldBeTrue)
		So(flag.Name, ShouldEqual, stringTestFlag.name)
		So(flag.Value, ShouldEqual, providedString)
		So(flag.Default, ShouldEqual, defaultString)
		valueFromMap, ok := flagMap[stringTestFlag.name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedString)

		// int
		flag, ok = flags[intTestFlag.name]
		So(ok, ShouldBeTrue)
		So(flag.Name, ShouldEqual, intTestFlag.name)
		So(fmt.Sprintf("%d", defaultInt), ShouldEqual, flag.Default)
		So(flag.Value, ShouldEqual, providedInt)
		valueFromMap, ok = flagMap[intTestFlag.name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedInt)

		// duration
		flag, ok = flags[durTestFlag.name]
		So(ok, ShouldBeTrue)
		So(durTestFlag.name, ShouldEqual, flag.Name)
		So(fmt.Sprintf("%s", defaultDuration), ShouldEqual, flag.Default)
		So(flag.Value, ShouldEqual, providedDuration)
		valueFromMap, ok = flagMap[durTestFlag.name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedDuration)

		// slice
		flag, ok = flags[sliceTestFlag.name]
		So(ok, ShouldBeTrue)
		So(sliceTestFlag.name, ShouldEqual, flag.Name)
		So(flag.Value, ShouldEqual, providedSlice)
		valueFromMap, ok = flagMap[sliceTestFlag.name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedSlice)

		Convey("Configuration file is also generted correctly", func() {

			body := DumpConfig()
			requriredParts := []string{
				"# stringDesc",
				"# Default: http://foo-bar",
				"SWAN_STRINGTEST=bar-foo",
				"# intDesc",
				"# Default: 628",
				"SWAN_INTTEST=13",
				"# durDesc",
				"# Default: 2m3s",
				"SWAN_DURATIONTEST=2h0m0s",
				"# sliceDesc",
				"SWAN_SLICETEST=foo1,foo2",
				"set +o allexport",
			}

			for _, part := range requriredParts {
				So(body, ShouldContainSubstring, part)
			}

			Convey("even with overwritten given values", func() {
				body := DumpConfigMap(map[string]string{
					"stringTest":   "newString",
					"intTest":      "17",
					"durationTest": "3h",
					"sliceTest":    "bar1,bar2",
				})
				requriredParts := []string{
					"SWAN_STRINGTEST=",
					"SWAN_INTTEST=17",
					"SWAN_DURATIONTEST=3h",
					"SWAN_SLICETEST=bar1,bar2",
				}

				for _, part := range requriredParts {
					So(body, ShouldContainSubstring, part)
				}

			})
		})
	})

}
