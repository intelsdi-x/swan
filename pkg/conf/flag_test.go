package conf

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnvFlag(t *testing.T) {
	Convey("While using Flag struct, it should construct proper swan environment var name", t, func() {
		So(envName("test_name"), ShouldEqual, "SWAN_TEST_NAME")
	})
}

func TestFlags(t *testing.T) {
	Convey("While using Conf flags", t, func() {
		Convey("When some custom String Flag is defined", func() {
			// Register custom flag.
			customFlag := NewStringFlag("custom_string_arg", "help", "default")
			So(customFlag.Value(), ShouldEqual, "default")

			ParseFlags()
			So(customFlag.Value(), ShouldEqual, "default")

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "customContent"
				os.Setenv(envName(customFlag.Name), customValue)

				ParseFlags()
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Int Flag is defined", func() {
			// Register custom flag.
			customFlag := NewIntFlag("custom_int_arg", "help", 23424)

			So(customFlag.Value(), ShouldEqual, 23424)

			ParseFlags()
			So(customFlag.Value(), ShouldEqual, 23424)

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := 12
				os.Setenv(envName(customFlag.Name), fmt.Sprintf("%d", customValue))

				ParseFlags()
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Slice Flag is defined", func() {
			// Register custom flag.
			customFlag := NewSliceFlag("custom_slice_arg", "help")

			So(customFlag.Value(), ShouldResemble, []string{})

			ParseFlags()
			So(customFlag.Value(), ShouldResemble, []string{})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "A,B,C"
				os.Setenv(envName(customFlag.Name), customValue)
				ParseFlags()
				So(customFlag.Value(), ShouldResemble, []string{"A", "B", "C"})
			})
		})

		Convey("When some custom Bool Flag is defined", func() {
			// Register custom flag.
			customFlag := NewBoolFlag("custom_bool_arg", "help", false)

			So(customFlag.Value(), ShouldEqual, false)

			ParseFlags()
			So(customFlag.Value(), ShouldEqual, false)

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := true
				os.Setenv(envName(customFlag.Name), fmt.Sprintf("%v", customValue))

				ParseFlags()
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Duration Flag is defined", func() {
			// Register custom flag.
			customFlag := NewDurationFlag("custom_duration_arg", "help", 99*time.Millisecond)

			So(customFlag.Value(), ShouldEqual, 99*time.Millisecond)

			ParseFlags()
			So(customFlag.Value(), ShouldEqual, 99*time.Millisecond)

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := 1234 * time.Second
				os.Setenv(envName(customFlag.Name), customValue.String())

				ParseFlags()
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom IntSet Flag is defined", func() {
			// Register custom flag.
			customFlag := NewIntSetFlag("custom_intset_arg", "help", "")

			So(customFlag.Value(), ShouldResemble, isolation.IntSet{})

			ParseFlags()
			So(customFlag.Value(), ShouldResemble, isolation.IntSet{})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				err := os.Setenv(envName(customFlag.Name), "1,3-4,7-8")
				So(err, ShouldBeNil)

				ParseFlags()
				So(customFlag.Value(), ShouldResemble, isolation.IntSet{1: {}, 3: {}, 4: {}, 7: {}, 8: {}})
			})
		})
	})
}

func TestConfiguration(t *testing.T) {
	Convey("While using flags, we can extract right values for different types.", t, func() {

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

		intSetTestFlag := NewIntSetFlag("intSetTest", "intSetDesc", "")
		providedIntSet := "1,3-5"
		providedIntSetFlat := "1,3,4,5"

		flag.CommandLine.Parse([]string{
			"--intTest", providedInt,
			"--durationTest", providedDuration,
			"--stringTest", providedString,
			"--sliceTest", providedSlice,
			"--intSetTest", providedIntSet,
		})

		// External interface (just returns current value by name).
		flagMap := GetFlags()

		// Gather configuration and put into map (for testing purposes).
		// Prepare map with all flags for easier assertions.
		flags := map[string]flag.Flag{}
		for _, flag := range getFlagsDefinition() {
			flags[flag.Name] = *flag
		}

		// string
		name := stringTestFlag.Name
		flag, ok := flags[name]
		So(ok, ShouldBeTrue)
		So(flag.Name, ShouldEqual, name)
		So(flag.Value.String(), ShouldEqual, providedString)
		So(flag.DefValue, ShouldEqual, defaultString)
		valueFromMap, ok := flagMap[name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedString)

		// int
		name = intTestFlag.Name
		flag, ok = flags[name]
		So(ok, ShouldBeTrue)
		So(flag.Name, ShouldEqual, name)
		So(fmt.Sprintf("%d", defaultInt), ShouldEqual, flag.DefValue)
		So(flag.Value.String(), ShouldEqual, providedInt)
		valueFromMap, ok = flagMap[name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedInt)

		// duration
		name = durTestFlag.Name
		flag, ok = flags[name]
		So(ok, ShouldBeTrue)
		So(name, ShouldEqual, flag.Name)
		So(fmt.Sprintf("%s", defaultDuration), ShouldEqual, flag.DefValue)
		So(flag.Value.String(), ShouldEqual, providedDuration)
		valueFromMap, ok = flagMap[name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedDuration)

		// slice
		name = sliceTestFlag.Name
		flag, ok = flags[name]
		So(ok, ShouldBeTrue)
		So(name, ShouldEqual, flag.Name)
		So(flag.Value.String(), ShouldEqual, providedSlice)
		valueFromMap, ok = flagMap[name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedSlice)

		// intSet
		name = intSetTestFlag.Name
		flag, ok = flags[name]
		So(ok, ShouldBeTrue)
		So(name, ShouldEqual, flag.Name)
		So(flag.Value.String(), ShouldEqual, providedIntSetFlat)
		valueFromMap, ok = flagMap[name]
		So(ok, ShouldBeTrue)
		So(valueFromMap, ShouldEqual, providedIntSetFlat)

		Convey("Configuration file is also generated correctly", func() {

			body := DumpConfig()
			requriredParts := []string{
				"# stringDesc",
				"# Default: http://foo-bar",
				"STRINGTEST=bar-foo",
				"# intDesc",
				"# Default: 628",
				"INTTEST=13",
				"# durDesc",
				"# Default: 2m3s",
				"DURATIONTEST=2h0m0s",
				"# sliceDesc",
				"SLICETEST=foo1,foo2",
				"# intSetDesc",
				"INTSETTEST=1,3,4,5",
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
					"intSetTest":   "1,3,4,5",
				})
				expectedParts := []string{
					"STRINGTEST=newString",
					"INTTEST=17",
					"DURATIONTEST=3h",
					"SLICETEST=bar1,bar2",
					"INTSETTEST=1,3,4,5",
				}

				for _, part := range expectedParts {
					So(body, ShouldContainSubstring, part)
				}

			})
		})
	})

}
