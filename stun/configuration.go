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
	ENV_PREFIX          = "LSTN"
	KEY_UDP_PORT        = "udp.port"
	KEY_TCP_PORT        = "tcp.port"
	KEY_MONITORING_PORT = "monitoring.port"
	KEY_MONITORING_PATH = "monitoring.path"
	FLAG_UDP_PORT       = "udp-port"
	FLAG_TCP_PORT       = "tcp-port"
)

// default values
const (
	DEFAULT_UDP_PORT        = 3478
	DEFAULT_TCP_PORT        = 3478
	DEFAULT_MONITORING_PORT = 8081
	DEFAULT_MONITORING_PATH = "/metrics"
)

type ServerConf struct {
	Enabled bool
	Port    int
}

func (self ServerConf) String() string {
	return fmt.Sprintf("{enabled: %t, Port: %d}", self.Enabled, self.Port)
}

type MonitoringConf struct {
	Port int
	Path string
}

func (self MonitoringConf) String() string {
	return fmt.Sprintf("{Port: %d, Path: %s}", self.Port, self.Path)
}

type Configuration struct {
	Udp        ServerConf
	Tcp        ServerConf
	Monitoring MonitoringConf
}

func (self Configuration) String() string {
	return fmt.Sprintf("{Udp: %s, Tcp: %s Monitoring: %s}", self.Udp.String(), self.Tcp.String(), self.Monitoring.String())
}

func GetConfiguration() (*Configuration, error) {
	// let viper read from configuration file
	viper.SetConfigName("stun")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/stun/")
	viper.AddConfigPath("./config/")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Reading config file %s failed with error: %s", viper.ConfigFileUsed(), err)
	}

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
	err = viper.BindEnv(KEY_MONITORING_PORT)
	if nil != err {
		log.Printf("Bind env failed for key %s with error: %s", KEY_MONITORING_PORT, err)
	}
	err = viper.BindEnv(KEY_MONITORING_PATH)
	if nil != err {
		log.Printf("Bind env failed for key %s with error: %s", KEY_MONITORING_PATH, err)
	}

	viper.SetDefault(KEY_MONITORING_PORT, DEFAULT_MONITORING_PORT)
	viper.SetDefault(KEY_MONITORING_PATH, DEFAULT_MONITORING_PATH)

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
