package apptest

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/pion/stun"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	stunPort = "3478"
)

type LogConsumer struct {
}

func (g *LogConsumer) Accept(l testcontainers.Log) {
	log.Println(string(l.Content))
}

func TestWithPion(t *testing.T) {
	ctx := context.Background()
	var logConsumer LogConsumer
	exposedPort := fmt.Sprintf("%s:%s/udp", stunPort, stunPort)
	req := testcontainers.ContainerRequest{
		Image:        "stun:latest",
		ExposedPorts: []string{exposedPort},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&logConsumer},
		},
		WaitingFor: wait.ForLog("Starting Stun server, listening port at 3478"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start app: %s", err)
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop app: %s", err)
		}
	}()

	// Parse a STUN URI
	u, err := stun.ParseURI("stun:localhost:" + stunPort)
	if err != nil {
		panic(err)
	}

	// Creating a "connection" to STUN server.
	c, err := stun.DialURI(u, &stun.DialConfig{})
	if err != nil {
		panic(err)
	}
	// Building binding request with random transaction id.
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	// Sending request to STUN server, waiting for response message.
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}
		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			log.Println("No xor mapped addr")
		}

		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var mappedAddr stun.MappedAddress
		if err := mappedAddr.GetFrom(res.Message); err != nil {
			log.Println("No mapped addr")
		}

		fmt.Println("your xor mapped IP is", xorAddr.IP)
		fmt.Println("your mapped IP is", mappedAddr.IP)

		time.Sleep(5 * time.Second)
	}); err != nil {
		panic(err)
	}
}
