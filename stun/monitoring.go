package stun

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MonitoringStart(ctx context.Context, conf MonitoringConf, wg *sync.WaitGroup) {

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
