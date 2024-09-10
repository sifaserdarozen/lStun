package stun

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerMetrics() {
	var (
		BuildInfo = prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "lstun",
				Name:      "build_info",
				Help:      "Information about lstun binary",
				ConstLabels: prometheus.Labels{
					"version":    Version,
					"build_date": BuildDate,
					"env":        Env,
				},
			},
			func() float64 { return 1 },
		)
		UptimeInfo = prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "lstun",
				Name:      "uptime_sec",
				Help:      "Information about binary uptime",
			},
			func() float64 { return time.Since(StartDate).Truncate(time.Second).Seconds() },
		)
	)

	prometheus.MustRegister(BuildInfo)
	prometheus.MustRegister(UptimeInfo)
}

func MonitoringStart(ctx context.Context, conf MonitoringConf, wg *sync.WaitGroup) {

	registerMetrics()
	addr := fmt.Sprintf(":%d", conf.Port)
	log.Println("Monitoring at: ", addr)
	srv := http.Server{Addr: addr}
	http.Handle(conf.Path, promhttp.Handler())
	(*wg).Add(1)
	go func() {
		defer (*wg).Done()
		log.Printf("Starting Monitoring server, listening port at %d/%s", conf.Port, conf.Path)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("Monitoring server listen error: %v", err)
		}
		log.Println("Stopped Monitoring server ...")
	}()

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		<-ctx.Done()
		log.Println("Stopping Monitoring server ...")

		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("Monitoring server shutdown error: %v", err)
		}
	}()
}
