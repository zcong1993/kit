package main

import (
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extapp"
)

var (
	app = &cobra.Command{
		Use:   "mono",
		Short: "Mono repo example",
	}
)

func main() {
	app.AddCommand(service1Cmd, service2Cmd)
	extapp.RunDefaultServerApp(app)
}
