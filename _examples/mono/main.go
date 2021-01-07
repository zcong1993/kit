package main

import (
	"log"

	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/tracing/register"
)

var (
	app = &cobra.Command{
		Use:   "mono",
		Short: "Mono repo example",
	}
)

func main() {
	app.AddCommand(service1Cmd, service2Cmd)
	// 注册日志相关 flag
	log2.Register(app.PersistentFlags())
	register.RegisterFlags(app.PersistentFlags())

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
