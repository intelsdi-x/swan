package conf

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/isolation"
)

// EnvironmentPrefix is prefix that is used for evironment based configuration.
const EnvironmentPrefix = "SWAN_"

// Registry of flag names (required for proper ordering and semantic grouping instead of lexicographical order).
// Note: only the flags registered by our wrappers will dumped/stored in metadata.
var flagNames []string

func registerName(name string) {
	flagNames = append(flagNames, name)
}

func envName(name string) string {
	return fmt.Sprintf("%s%s", EnvironmentPrefix, strings.ToUpper(name))
}

// Flag represents option's definition from CLI and Environment variable.
type Flag struct {
	Name  string
	usage string
}

// StringFlag ...
type StringFlag struct {
	Flag
	value *string
}

// NewStringFlag is a constructor of StringFlag struct.
func NewStringFlag(name string, usage string, value string) StringFlag {
	registerName(name)
	return StringFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: flag.String(name, value, usage),
	}
}

// Value returns value of defined flag after parse.
func (s StringFlag) Value() string {
	return *s.value
}

// IntFlag represents flag with int value.
type IntFlag struct {
	Flag
	value *int
}

// NewIntFlag is a constructor of IntFlag struct.
func NewIntFlag(name string, usage string, value int) IntFlag {
	registerName(name)
	return IntFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: flag.Int(name, value, usage),
	}

}

// Value returns value of defined flag after parse.
func (i IntFlag) Value() int {
	return *i.value
}

// SliceFlag represents flag with slice string values.
type SliceFlag struct {
	Flag
	value *string // stored as csv.
}

// NewSliceFlag is a constructor of SliceFlag struct.
func NewSliceFlag(name string, usage string, value []string) SliceFlag {
	registerName(name)
	return SliceFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: flag.String(name, strings.Join(value, ","), usage),
	}
}

// Value returns value of defined flag after parse.
func (s SliceFlag) Value() []string {
	if *s.value == "" {
		return []string{}
	}

	return strings.Split(*s.value, ",")
}

// BoolFlag represents flag with bool value.
type BoolFlag struct {
	Flag
	value *bool
}

// NewBoolFlag is a constructor of BoolFlag struct.
func NewBoolFlag(name string, usage string, value bool) BoolFlag {
	registerName(name)
	return BoolFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: flag.Bool(name, value, usage),
	}
}

// Value returns value of defined flag after parse.
func (b BoolFlag) Value() bool {
	return *b.value
}

// DurationFlag represents flag with duration value.
type DurationFlag struct {
	Flag
	value *time.Duration
}

// NewDurationFlag is a constructor of DurationFlag struct.
func NewDurationFlag(name string, usage string, value time.Duration) DurationFlag {
	registerName(name)
	return DurationFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: flag.Duration(name, value, usage),
	}
}

// Value returns value of defined flag after parse.
func (d DurationFlag) Value() time.Duration {
	return *d.value
}

// IntSetFlag represents flag with set of integers value.
type IntSetFlag struct {
	Flag
	value *isolation.IntSet
}

// NewIntSetFlag is a constructor of string range based flags
func NewIntSetFlag(name string, usage string, value string) *IntSetFlag {
	registerName(name)
	intSet, err := isolation.NewIntSetFromRange(value)
	if err != nil {
		panic(fmt.Errorf("invalid default value for IntSetFlag: %q", value))
	}
	intSetFlag := &IntSetFlag{
		Flag: Flag{
			Name:  name,
			usage: usage,
		},
		value: &intSet,
	}
	flag.Var(intSetFlag, name, usage)
	return intSetFlag
}

// Value returns value of defined flag after parse.
func (isf *IntSetFlag) Value() isolation.IntSet {
	return *isf.value
}

// Set new value for IntSetFlag.
func (isf *IntSetFlag) Set(value string) error {
	intSet, err := isolation.NewIntSetFromRange(value)
	if err != nil {
		return err
	}
	isf.value = &intSet
	return nil
}

// String method returns IntSet as string.
func (isf *IntSetFlag) String() string {
	// Required because Internally go flag package creates zero value instance.
	if isf.value != nil {
		return isf.value.AsRangeString()
	}
	return ""
}
