package main

import (
	"context"
	"log"
	"sync"

	"github.com/sifaserdarozen/stun/stun"
)

func main() {
	// read configuration
	conf := stun.GetConfiguration()
	log.Println(conf)

	// sync helpers
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	// start stun service
	stun.Start(conf, ctx, &wg)

	// wait till softkill
	stun.WaitTillInterrupt()

	// cancel context
	cancel()

	// wait for threads to finish
	wg.Wait()
}
