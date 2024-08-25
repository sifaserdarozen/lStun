package stun

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// viper keys
const (
	ENV_PREFIX    = "LSTN"
	KEY_UDP_PORT  = "udp.port"
	KEY_TCP_PORT  = "tcp.port"
	FLAG_UDP_PORT = "udp-port"
	FLAG_TCP_PORT = "tcp-port"
)

// default values
const (
	DEFAULT_UDP_PORT = 3478
	DEFAULT_TCP_PORT = 3478
)

type ServerConf struct {
	Enabled bool
	Port    int
}

func (self ServerConf) String() string {
	return fmt.Sprintf("{enabled: %t, Port: %d}", self.Enabled, self.Port)
}

type Configuration struct {
	Udp ServerConf
	Tcp ServerConf
}

func (self Configuration) String() string {
	return fmt.Sprintf("{Udp: %s, Tcp: %s}", self.Udp.String(), self.Tcp.String())
}

func GetConfiguration() (*Configuration, error) {

	// let viper set environment variables prefix and register keys to look for
	viper.SetEnvPrefix(ENV_PREFIX)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.BindEnv(KEY_UDP_PORT)
	if nil != err {
		log.Printf("Bind env failed for key %s with error: %s", KEY_UDP_PORT, err)
	}
	err = viper.BindEnv(KEY_TCP_PORT)
	if nil != err {
		log.Printf("Bind env failed for key %s with error: %s", KEY_TCP_PORT, err)
	}

	// use golang flag to get cli argumenst
	flag.Int(FLAG_UDP_PORT, DEFAULT_UDP_PORT, "Stun server udp port")
	flag.Int(FLAG_TCP_PORT, DEFAULT_TCP_PORT, "Stun server tcp port")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// let viper read from flags (CLI)
	err = viper.BindPFlag(KEY_UDP_PORT, pflag.Lookup(FLAG_UDP_PORT))
	if nil != err {
		log.Printf("Bind flag failed for key %s with error: %s", KEY_UDP_PORT, err)
	}
	err = viper.BindPFlag(KEY_TCP_PORT, pflag.Lookup(FLAG_TCP_PORT))
	if nil != err {
		log.Printf("Bind flag failed for key %s with error: %s", KEY_TCP_PORT, err)
	}

	config := Configuration{}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("Error in unmarshalling configuration, %s", err)
		return nil, err
	}

	log.Println("Using configuration...")
	log.Println(config)

	return &config, nil
}
