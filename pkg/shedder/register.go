package shedder

import (
	"github.com/spf13/cobra"
	"github.com/tal-tech/go-zero/core/load"
)

const (
	shedderCpuThreshold = "shedder.cpu-threshold"
	helpText            = "CpuThreshold config for adaptive shedder, set 0 to disable"
)

// Factory is shedder factory.
type Factory = func() load.Shedder

// Register register the shedder option into cobra global flag set.
func Register(app *cobra.Command) Factory {
	var cpuThreshold int64

	app.PersistentFlags().Int64Var(&cpuThreshold, shedderCpuThreshold, 900, helpText)

	return func() load.Shedder {
		return NewShedder(cpuThreshold)
	}
}

// NewShedder create a new shedder instance.
func NewShedder(cpuThreshold int64) load.Shedder {
	if cpuThreshold == 0 {
		return nil
	}
	return load.NewAdaptiveShedder(load.WithCpuThreshold(cpuThreshold))
}
