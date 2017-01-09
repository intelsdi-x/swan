package conf

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

const stringListDelimiter = ","

// StringListValue is a custom kingpin parser which resolves flag's parameters which consists of
// string slice delimited by `stringListDelimiter`.
// For instance for delimiter = "," and flag defined like this:
// `flag = StringList(kingpin.Flag("flag_name", "help").Short("f"))`
//
// When user would specify options: `-f=A,B,C -f=D,E,F` our `flag` variable would be a slice with
// A,B,C,D,E,F items.
type StringListValue []string

// Set parses the input string and append that as a slice. Implements kingpin.Value.
func (s *StringListValue) Set(value string) error {
	// Split string from input to slice and merge with saved slice.
	*s = append(*s, strings.Split(value, stringListDelimiter)...)
	// TODO(bp): Remove duplicates?
	return nil
}

// String returns string value from StringListVar. Implements kingpin.Value.
func (s *StringListValue) String() string {
	return fmt.Sprintf("%v", ([]string)(*s))
}

// Get retrieves content of StringListVar. Implements kingpin.Getter.
func (s *StringListValue) Get() interface{} {
	return ([]string)(*s)
}

// IsCumulative implements optional interface (kingpin.repeatableFlag) for flags that can be repeated.
func (s *StringListValue) IsCumulative() bool {
	return true
}

// StringList is a helper for defining kingping flags.
func StringList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*StringListValue)(target))
	return
}
