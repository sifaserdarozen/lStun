package stun

import (
	"flag"
	"fmt"
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
	udpPort := flag.Int("udpPort", UDP_PORT, "Stun server udp port")
	tcpPort := flag.Int("tcpPort", TCP_PORT, "Stun server tcp port")

	flag.Parse()

	// Shall we also check environment variables for configuration?
	// In that case it will be better to set priority to something like
	// Arguments > Environment variables > Defaults

	return Configuration{
		udpPort: *udpPort,
		tcpPort: *tcpPort,
	}
}
