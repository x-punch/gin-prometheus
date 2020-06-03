package main

import (
	"time"

	"github.com/gin-gonic/gin"
	prometheus "github.com/x-punch/gin-prometheus"
)

func main() {
	app := gin.New()
	app.Use(gin.Logger())
	prom := prometheus.NewPrometheus("gin")
	prom.UseWithAuth(app, gin.Accounts{"gin": "gonic"}, "/metrics")

	app.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"package": "github.com/x-punch/gin-prometheus"})
	})
	app.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": "1.0.0", "time": time.Now()})
	})
	app.POST("/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	if err := app.Run(":80"); err != nil {
		panic(err)
	}
}
