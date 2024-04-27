package string_slice

import (
	"strings"
)

type StringSlice []string

// Convert xs to a string.  This is needed for the flag.Value
// interface so we can read comman-separated values into a
// StringSlice.
func (xs StringSlice) String() string {
	var result strings.Builder
	for i, x := range xs {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(x)
	}
	return result.String()
}

// Set the value of s to xs.  To specify multiple values separate each
// with a comma and no space.  This is needed for the flag.Value
// interface so we can read comman-separated values into xs.
func (xs *StringSlice) Set(s string) error {
	for _, part := range strings.Split(s, ",") {
		*xs = append(*xs, strings.TrimSpace(part))
	}
	return nil
}
