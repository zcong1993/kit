package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"

	"github.com/oklog/run"
	"github.com/zcong1993/x/pkg/extrun"

	"github.com/gin-gonic/gin"
	"github.com/zcong1993/x/pkg/ginhelper"
)

type Input struct {
	Name string `json:"name" binding:"required"`
}

func main() {
	app := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			logger := log2.MustNewLogger(cmd)
			r := ginhelper.DefaultWithLogger(logger)
			r.GET("/", func(c *gin.Context) {
				time.Sleep(5 * time.Second)
				c.String(http.StatusOK, "Welcome Gin Server")
			})

			r.GET("/test", ginhelper.ErrorWrapper(func(c *gin.Context) error {
				q := c.Query("q")
				if q == "" {
					return ginhelper.NewBizError(400, 400, "query q is required")
				}
				c.JSON(200, gin.H{"success": true})
				return nil
			}))

			r.POST("/p", ginhelper.ErrorWrapper(func(c *gin.Context) error {
				var input Input
				err := c.ShouldBind(&input)
				if err != nil {
					fmt.Printf("%+v\n", err)
					return err
				}
				c.JSON(200, &input)
				return nil
			}))

			httpServer := ginhelper.NewHttpServer(r)

			var g run.Group
			extrun.HandleSignal(&g)

			g.Add(func() error {
				return httpServer.Start(":8080")
			}, func(err error) {
				fmt.Println("http server will exit", err)
				_ = httpServer.Shutdown(time.Second * 5)
			})

			if err := g.Run(); err != nil {
				log.Fatal("start error ", err)
			}
		},
	}

	log2.Register(app)

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
