package main

import (
	"fmt"

	"github.com/go-kit/kit/log/level"
	"github.com/zcong1993/x/pkg/log/flag"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		positionFile = kingpin.Flag("position.file", "Position file name.").Default("position.json").String()
		interval     = kingpin.Flag("interval", "All ticker interval.").Default("10s").Duration()
	)
	loggerFunc := flag.NewFactoryFromFlags(kingpin.CommandLine)
	interval2 := kingpin.Flag("interval2", "All ticker interval.").Default("12s").Duration()
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := loggerFunc()

	fmt.Println(*positionFile, *interval, *interval2)
	logger.Log("aaa", "xcdcd")
	level.Debug(logger).Log("msg", "debug")
}
