# gin-prometheus
Gin middleware used to export prometheus metrics

# Usage
```go
app := gin.New()
prometheus := prometheus.NewPrometheus("service")
prometheus.Use(app, "/metrics")
```

# Auth Example
```go
package main

import (
	"github.com/gin-gonic/gin"
	prometheus "github.com/x-punch/gin-prometheus"
)

func main() {
	app := gin.New()
	app.Use(gin.Logger())
	prom := prometheus.NewPrometheus("gin")
	prom.UseWithAuth(app, gin.Accounts{"gin": "gonic"}, "/metrics")
	if err := app.Run(":80"); err != nil {
		panic(err)
	}
}
```