package snclient

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	AvailableListeners = append(AvailableListeners, ListenHandler{"PrometheusServer", "/settings/Prometheus/server", NewHandlerPrometheus()})
}

var (
	prometheusRegistered bool

	infoCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "snclient_info",
			Help: "information about this agent",
		},
		[]string{"version", "build"})
)

type HandlerPrometheus struct {
	noCopy  noCopy
	handler http.Handler
}

func NewHandlerPrometheus() *HandlerPrometheus {
	handler := &HandlerPrometheus{
		handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer,
			promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{EnableOpenMetrics: true},
			),
		),
	}

	return handler
}

func (l *HandlerPrometheus) Type() string {
	return "prometheus"
}

func (l *HandlerPrometheus) Defaults() ConfigSection {
	defaults := ConfigSection{
		"port":    "9999",
		"use ssl": "0",
	}
	defaults.Merge(DefaultListenHTTPConfig)

	return defaults
}

func (l *HandlerPrometheus) Init(snc *Agent) error {
	registerMetrics()
	infoCount.WithLabelValues(VERSION, snc.Build).Set(1)

	return nil
}

func (l *HandlerPrometheus) GetMappings(*Agent) []URLMapping {
	return []URLMapping{
		{URL: "/metrics", Handler: l.handler},
	}
}

func registerMetrics() {
	// registering twice will throw lots of errors
	if prometheusRegistered {
		return
	}

	prometheusRegistered = true

	// register the metrics
	if err := prometheus.Register(infoCount); err != nil {
		log.Errorf("failed to register prometheus metric: %s", err.Error())
	}
}