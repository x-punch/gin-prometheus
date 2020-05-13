package prometheus

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// URLFormater used to format request url, it can be used to replace url params or remove raw query
type URLFormater func(c *gin.Context) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	uptime               *prometheus.CounterVec
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz prometheus.Summary
	router               *gin.Engine
	listenAddress        string
	PushGateway          PushGateway
	MetricsList          []*Metric
	FormatCounterURL     URLFormater
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string) *Prometheus {
	var metricsList []*Metric
	for _, metric := range standardMetrics {
		metricsList = append(metricsList, metric)
	}

	p := &Prometheus{
		MetricsList: metricsList,
		FormatCounterURL: func(c *gin.Context) string {
			url := c.Request.URL.EscapedPath()
			for _, p := range c.Params {
				url = strings.Replace(url, p.Value, ":"+p.Key, 1)
			}
			return url
		},
	}
	p.registerMetrics(subsystem)
	for range time.Tick(time.Second) {
		p.uptime.WithLabelValues().Inc()
	}

	return p
}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine, metricsPath string) {
	e.Use(p.handlerFunc(metricsPath))
	p.setMetricsPath(e, metricsPath)
}

// UseWithAuth adds the middleware to a gin engine with BasicAuth.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts, metricsPath string) {
	e.Use(p.handlerFunc(metricsPath))
	p.setMetricsPathWithAuth(e, accounts, metricsPath)
}

// UseWithCustomMetrics use the custom metrics
func (p *Prometheus) UseWithCustomMetrics(e *gin.Engine, gatherer prometheus.Gatherers, metricsPath string) {
	p.setMetricsPathWithCustomMetrics(e, gatherer, metricsPath)
}

// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
// same address of the gin engine that is being used
func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

// SetListenAddressWithRouter for using a separate router to expose metrics. (this keeps things like GET /metrics out of
// your content's access log).
func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
}

func (p *Prometheus) setMetricsPath(e *gin.Engine, metricsPath string) {
	if p.listenAddress != "" {
		p.router.GET(metricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(metricsPath, prometheusHandler())
	}
}

func (p *Prometheus) setMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts, metricsPath string) {
	if p.listenAddress != "" {
		p.router.GET(metricsPath, gin.BasicAuth(accounts), prometheusHandler())
		p.runServer()
	} else {
		e.GET(metricsPath, gin.BasicAuth(accounts), prometheusHandler())
	}
}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go func() {
			if err := p.router.Run(p.listenAddress); err != nil {
				errorLog(err.Error())
			}
		}()
	}
}

func (p *Prometheus) setMetricsPathWithCustomMetrics(e *gin.Engine, gatherer prometheus.Gatherers, metricsPath string) {
	p.router.GET(metricsPath, prometheusHandlerFor(gatherer))
}

func (p *Prometheus) handlerFunc(metricsPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == metricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		p.reqDur.Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, c.Request.Host, p.FormatCounterURL(c)).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func prometheusHandlerFor(gatherer prometheus.Gatherers) gin.HandlerFunc {
	h := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	var s int
	if r.URL != nil {
		s = len(r.URL.String())
	}
	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)
	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

func errorLog(err string) {
	log.Print("[PROM]" + err)
}
