package dcompress

import "fmt"

type VerboseType bool

var (
	Verbose VerboseType
)

func (v VerboseType) Printf(s string, a ...interface{}) {
	if v {
		fmt.Printf(s, a...)
	}
}
