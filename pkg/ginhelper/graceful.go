package ginhelper

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	engine *gin.Engine
	srv    *http.Server
}

func NewHttpServer(engine *gin.Engine) *HttpServer {
	return &HttpServer{engine: engine}
}

func (hs *HttpServer) Start(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: hs.engine,
	}
	hs.srv = srv
	return srv.ListenAndServe()
}

func (hs *HttpServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return hs.srv.Shutdown(ctx)
}
