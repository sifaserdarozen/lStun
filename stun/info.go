package stun

import (
	"log"
	"net"
	"os"
	"time"
)

const (
	ENV_KEY          = "LSTN_INFO_ENV"
	DEFAULT_INFO_ENV = "local"
)

var (
	// version & build date gets defined by the build system
	Version   = "dirty"
	BuildDate string

	Env       = DEFAULT_INFO_ENV
	StartDate = time.Now()
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

func getEnv() {
	val, ok := os.LookupEnv(ENV_KEY)
	if !ok {
		Env = DEFAULT_INFO_ENV
	} else {
		Env = val
	}
}

func InitInfo() {
	getItfcs()
	getEnv()
}
