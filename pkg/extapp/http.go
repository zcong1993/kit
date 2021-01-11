package extapp

import (
	"log"

	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/shedder"
	"github.com/zcong1993/x/pkg/tracing/register"
)

func RunDefaultHttpServerApp(app *cobra.Command) {
	// 注册日志相关 flag
	log2.Register(app.PersistentFlags())
	// 注册 tracing flag
	register.RegisterFlags(app.PersistentFlags())
	// 注册 shedder flag
	shedder.Register(app.PersistentFlags())

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
