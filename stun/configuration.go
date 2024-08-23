package stun

import (
	"flag"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// default values
const (
	UDP_PORT = 3478
	TCP_PORT = 3478
)

type Configuration struct {
	udpPort int
	tcpPort int
}

func (self Configuration) String() string {
	return fmt.Sprintf("{Udp Port: %d, Tcp Port: %d}", self.udpPort, self.tcpPort)
}

func GetConfiguration() Configuration {
	flag.Int("udpPort", UDP_PORT, "Stun server udp port")
	flag.Int("tcpPort", TCP_PORT, "Stun server tcp port")

	// let viper read from flags (CLI)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Shall we also check environment variables for configuration?
	// In that case it will be better to set priority to something like
	// CLI/flags > Environment variables > Configuration file > Defaults

	return Configuration{
		udpPort: viper.GetInt("udpPort"),
		tcpPort: viper.GetInt("tcpPort"),
	}
}
