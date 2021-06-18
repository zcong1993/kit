package main

import (
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extapp"
)

var (
	cmd = &cobra.Command{
		Use:   "mono",
		Short: "Mono repo example",
	}
)

func main() {
	app := extapp.NewApp()

	cmd.AddCommand(service1Cmd(app), service2Cmd(app))
	app.RunDefaultServerApp(cmd)
}
