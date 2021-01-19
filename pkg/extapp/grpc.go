package extapp

import (
	"log"

	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/tracing/register"
)

func RunDefaultGrpcServerApp(app *cobra.Command) {
	// 注册日志相关 flag
	log2.Register(app.PersistentFlags())
	// 注册 tracing flag
	register.RegisterFlags(app.PersistentFlags())

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
