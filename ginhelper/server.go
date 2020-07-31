package ginhelper

import (
	"github.com/gin-gonic/gin"
	logKit "github.com/go-kit/kit/log"
	"github.com/zcong1993/x/log"
	"github.com/zcong1993/x/log/flag"
	"gopkg.in/alecthomas/kingpin.v2"
)

func DefaultServer() (*gin.Engine, logKit.Logger) {
	var logConfig = log.Config{}
	flag.AddFlags(kingpin.CommandLine, &logConfig)
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := log.New(&logConfig)
	r := DefaultWithLogger(logger)
	return r, logger
}
