package main

import (
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extapp"
)

var app = &cobra.Command{
	Use:   "grpc",
	Short: "Grpc mono repo example",
}

func main() {
	app.AddCommand(serviceCmd, gatewayCmd)

	extapp.RunDefaultServerApp(app)
}
