package shedder

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tal-tech/go-zero/core/load"
)

const (
	shedderCpuThreshold = "shedder.cpu-threshold"
	helpText            = "CpuThreshold config for adaptive shedder, set 0 to disable"
)

func Register(flagSet *pflag.FlagSet) {
	flagSet.Int64(shedderCpuThreshold, 800, helpText)
}

func NewShedder(cpuThreshold int64) load.Shedder {
	if cpuThreshold == 0 {
		return nil
	}
	return load.NewAdaptiveShedder(load.WithCpuThreshold(cpuThreshold))
}

func NewShedderFromCmd(cmd *cobra.Command) load.Shedder {
	cpuThreshold, err := cmd.Flags().GetInt64(shedderCpuThreshold)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed init adaptive shedder:  %s\n", err)
		os.Exit(2)
	}
	return NewShedder(cpuThreshold)
}
