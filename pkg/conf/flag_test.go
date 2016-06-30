package conf

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path"
	"testing"
	"time"
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

			Convey("When we not defined any environment variable we should have default value after parse", func() {
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

		Convey("When some custom File Flag is defined", func() {
			// Register custom flag.
			defaultFilePath := path.Join(fs.GetSwanPath(), "README.md")
			customFlag := NewFileFlag("custom_file_arg", "help", defaultFilePath)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, defaultFilePath)
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom existing path we should have that value after parse", func() {
				customValue := path.Join(fs.GetSwanPath(), "Makefile")
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})

			Convey("When we define custom not existing path we should have error after parse", func() {
				customValue := "non-existing/path"
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith, "does not exist")
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

			Convey("When we not defined any environment variable we should have default value after parse", func() {
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

			Convey("When we not defined any environment variable we should have default value after parse", func() {
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

			Convey("When we not defined any environment variable we should have default value after parse", func() {
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

			Convey("When we not defined any environment variable we should have default value after parse", func() {
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

		Convey("When some custom IP Flag is defined", func() {
			// Register custom flag.
			customFlag :=
				NewIPFlag("custom_ip_arg", "help", "255.255.255.5")
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, "255.255.255.5")
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldResemble, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "255.255.255.99"
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldResemble, customValue)
			})

			Convey("When we define invalid custom environment variable we should have error after parse", func() {
				Convey("Like too long IP", func() {
					invalidCustomValue := "255.255.255.99.324"
					os.Setenv(customFlag.envName(), invalidCustomValue)
				})

				Convey("Like IP with invalid characters", func() {
					invalidCustomValue := "255.dfg.255.99"
					os.Setenv(customFlag.envName(), invalidCustomValue)
				})

				Convey("Like IP with number above 255", func() {
					invalidCustomValue := "300.255.255.99"
					os.Setenv(customFlag.envName(), invalidCustomValue)
				})

				err := ParseEnv()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith, "is not an IP address")
			})
		})
	})
}
