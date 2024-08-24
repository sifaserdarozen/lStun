package stun

import (
	"flag"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// viper keys
const (
	ENV_PREFIX   = "LSTN"
	ENV_UDP_PORT = "LSTN_UDP_PORT"
	ENV_TCP_PORT = "LSTN_TCP_PORT"
	UDP_PORT     = "udpPort"
	TCP_PORT     = "tcpPort"
)

// default values
const (
	DEFAULT_UDP_PORT = 3478
	DEFAULT_TCP_PORT = 3478
)

type Configuration struct {
	udpPort int
	tcpPort int
}

func (self Configuration) String() string {
	return fmt.Sprintf("{Udp Port: %d, Tcp Port: %d}", self.udpPort, self.tcpPort)
}

func GetConfiguration() Configuration {
	// let viper set environment variables prefix and register keys to look for
	viper.SetEnvPrefix(ENV_PREFIX)
	err := viper.BindEnv(UDP_PORT, ENV_UDP_PORT)
	if nil != err {
		log.Printf("Env get failed for %s with error %s", ENV_UDP_PORT, err)
	}
	err = viper.BindEnv(TCP_PORT, ENV_TCP_PORT)
	if nil != err {
		log.Printf("Env get failed for %s with error: %s", ENV_TCP_PORT, err)
	}

	// use golang flag to get cli argumenst
	flag.Int(UDP_PORT, DEFAULT_UDP_PORT, "Stun server udp port")
	flag.Int(TCP_PORT, DEFAULT_TCP_PORT, "Stun server tcp port")

	// let viper read from flags (CLI)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err = viper.BindPFlags(pflag.CommandLine)
	if nil != err {
		log.Printf("Binfing flags failed with error: %s", err)
	}

	return Configuration{
		udpPort: viper.GetInt(UDP_PORT),
		tcpPort: viper.GetInt(TCP_PORT),
	}
}
