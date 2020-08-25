package ginhelper

import (
	"github.com/gin-gonic/gin"
	logKit "github.com/go-kit/kit/log"
	"github.com/zcong1993/x/log/flag"
	"gopkg.in/alecthomas/kingpin.v2"
)

func DefaultServer() (*gin.Engine, logKit.Logger) {
	loggerF := flag.NewFactoryFromFlags(kingpin.CommandLine)
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := loggerF()
	r := DefaultWithLogger(logger)
	return r, logger
}
