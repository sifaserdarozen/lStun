package stun

import (
	"log"
	"os"
	"os/signal"
)

func WaitTillInterrupt() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	<-signalChan
	log.Println("Stoping due to soft kill (kill -SIGINT <pid>), stopping threads ...")
}
