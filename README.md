# gin-prometheus
Gin middleware used to export prometheus metrics

# Usage
```
app := gin.New()
prometheus := middleware.NewPrometheus("service")
prometheus.Use(app, "/metrics")
```

# Auth
```
app := gin.New()
prometheus := middleware.NewPrometheus("service")
			prom.UseWithAuth(app, gin.Accounts{"usr": "pwd"}, "/metrics")
```