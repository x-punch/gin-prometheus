package prometheus

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// PushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PushGateway struct {
	// Push interval
	PushInterval time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	URL string

	// Local metrics URL where metrics are fetched from, this could be ommited in the future
	// if implemented using prometheus common/expfmt instead
	MetricsURL string

	// pushgateway job name, defaults to "gin"
	Job string
}

// SetPushGateway sends metrics to a remote pushgateway exposed on URL
// every PushInterval. Metrics are fetched from metricsURL
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushInterval time.Duration) {
	p.PushGateway.URL = pushGatewayURL
	p.PushGateway.MetricsURL = metricsURL
	p.PushGateway.PushInterval = pushInterval
	p.startPushTicker()
}

// SetPushGatewayJob job name, defaults to "gin"
func (p *Prometheus) SetPushGatewayJob(j string) {
	p.PushGateway.Job = j
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getURL(), bytes.NewBuffer(metrics))
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		p.Logger.Error("Error sending to push gateway: %s" + err.Error())
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(p.PushGateway.PushInterval)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

func (p *Prometheus) getMetrics() []byte {
	response, err := http.Get(p.PushGateway.MetricsURL)
	if err != nil {
		p.Logger.Error(err.Error())
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			p.Logger.Error(err.Error())
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		p.Logger.Error(err.Error())
	}

	return body
}

func (p *Prometheus) getURL() string {
	h, _ := os.Hostname()
	if p.PushGateway.Job == "" {
		p.PushGateway.Job = "gin"
	}
	return p.PushGateway.URL + "/metrics/job/" + p.PushGateway.Job + "/instance/" + h
}
