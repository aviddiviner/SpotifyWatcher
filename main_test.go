package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aviddiviner/docopt-go"
)

// List of test cases
var usageTestTable = []struct {
	argv string  // Given command line args
	opts options // Expected options parsed
}{
	{
		"-s 5 -t 3 -w6 -n 7 -f -v",
		options{
			TopInterval:     5,
			CpuThreshold:    3.0,
			WindowLength:    6,
			AllowedBreaches: 7,
			Quiet:           false,
			Force:           true,
			Verbose:         true,
		},
	},
}

func TestUsage(t *testing.T) {
	for _, tt := range usageTestTable {
		docopt.DefaultParser = &docopt.Parser{HelpHandler: func(err error, usage string) {
			t.Errorf("invalid usage: %s\n", tt.argv)
		}}
		opts := parseOptions(strings.Split(tt.argv, " "))
		if !reflect.DeepEqual(opts, tt.opts) {
			t.Errorf("result (1) doesn't match expected (2) \n%v \n%v", opts, tt.opts)
		}
	}
}
