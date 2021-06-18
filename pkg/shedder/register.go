package shedder

import (
	"github.com/spf13/cobra"
	"github.com/tal-tech/go-zero/core/load"
)

const (
	shedderCpuThreshold = "shedder.cpu-threshold"
	helpText            = "CpuThreshold config for adaptive shedder, set 0 to disable"
)

type Factory = func() load.Shedder

func Register(app *cobra.Command) Factory {
	var cpuThreshold int64

	app.PersistentFlags().Int64Var(&cpuThreshold, shedderCpuThreshold, 0, helpText)

	return func() load.Shedder {
		return NewShedder(cpuThreshold)
	}
}

func NewShedder(cpuThreshold int64) load.Shedder {
	if cpuThreshold == 0 {
		return nil
	}
	return load.NewAdaptiveShedder(load.WithCpuThreshold(cpuThreshold))
}
