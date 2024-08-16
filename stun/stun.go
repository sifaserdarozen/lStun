package stun

import (
	"context"
	"log"
	"net"
	"sync"
)

func getItfcs() {
	addrs, err := net.InterfaceAddrs()
	if nil != err {
		log.Println("error getting interfaces: ", err)
		return
	}

	log.Println("System Interfaces")
	for _, v := range addrs {
		log.Println("net: ", v.Network(), " addrs: ", v.String())
	}
}

func Start(conf Configuration, ctx context.Context, wg *sync.WaitGroup) {
	getItfcs()
	UdpStart(ctx, conf, wg)
	TcpStart(ctx, conf, wg)
}
