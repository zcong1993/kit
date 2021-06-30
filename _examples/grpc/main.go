package main

import (
	"github.com/spf13/cobra"
	"github.com/zcong1993/kit/pkg/extapp"
)

var cmd = &cobra.Command{
	Use:   "grpc",
	Short: "Grpc mono repo example",
}

func main() {
	app := extapp.NewApp()

	cmd.AddCommand(serviceCmd(app), gatewayCmd(app), middleCmd(app))

	app.Run(cmd)
}
