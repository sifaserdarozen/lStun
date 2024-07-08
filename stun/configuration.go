package stun

import (
	"flag"
	"fmt"
)

// default values
const (
	UDP_PORT = 3478
)

type Configuration struct {
	udpPort int
}

func (self Configuration) String() string {
	return fmt.Sprintf("Udp Port = %d", self.udpPort)
}

func GetConfiguration() Configuration {
	udpPort := flag.Int("udpPort", UDP_PORT, "Stun server udp listening port")

	flag.Parse()

	// Shall we also check environment variables for configuration?
	// In that case it will be better to set priority to something like
	// Arguments > Environment variables > Defaults

	return Configuration{
		udpPort: *udpPort,
	}
}
