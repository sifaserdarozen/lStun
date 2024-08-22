package stun

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func WaitTillInterrupt() {
	signalChan := make(chan os.Signal, 1)
	// register os generic os.Interrupt / os.Kill. Refer https://pkg.go.dev/os#Signal
	signal.Notify(signalChan, os.Interrupt)
	// register os specific syscall.SIGTERM, syscall.SIGHUP, etc
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signalChan)

loop:
	for {
		switch s := <-signalChan; s {
		case os.Interrupt:
			log.Println("Stoping due to soft kill (kill -SIGINT <pid>)")
			break loop
		case syscall.SIGTERM:
			log.Println("Stoping due to soft kill (kill -SIGTERM <pid>)")
			break loop
		default:
			// Implement SIGHUP for configuration reread
			log.Printf("Ignoring registered but unimplemented signal %T, %v", s, s)
		}
	}

	log.Println("Stopping threads ...")
}
