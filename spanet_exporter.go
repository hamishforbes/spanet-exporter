package main

import (
	"net/http"
	"os"

	"github.com/go-kit/log/level"
	"github.com/hamishforbes/spanet_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	username      = kingpin.Flag("spanet.username", "SpaNet Username.").Required().String()
	passwordHash  = kingpin.Flag("spanet.password-hash", "SpaNet password Hash.").Required().String()
	spaName       = kingpin.Flag("spanet.spaName", "SpaNet spa name.").Required().String()
	webConfig     = webflag.AddFlags(kingpin.CommandLine)
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9150").String()
	metricPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
)

func main() {
	// Parse flags.
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("spanet_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting spanet_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	prometheus.MustRegister(version.NewCollector("spanet_exporter"))
	prometheus.MustRegister(collector.New(*username, *passwordHash, *spaName, logger))

	http.Handle(*metricPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>SpaNet Exporter</title></head>
             <body>
             <h1>SpaNet Exporter</h1>
             <p><a href='` + *metricPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	srv := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
