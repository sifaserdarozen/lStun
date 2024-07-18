package apptest

import (
	"context"
	"fmt"
	"log"
	"testing"
	"net"

	"github.com/pion/stun/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/docker/go-connections/nat"
	dc "github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
)

const (
	stunPort = "3478"
	proto = "udp"
)

type LogConsumer struct {
}

func (g *LogConsumer) Accept(l testcontainers.Log) {
	log.Println(string(l.Content))
}

func TestWithPion(t *testing.T) {
	ctx := context.Background()
	var logConsumer LogConsumer

	// docker run -p 127.0.0.1::3478/udp stun:latest
	exposedPort, err := nat.NewPort(proto, stunPort)
	if err != nil {
		t.Error(err)
	}
	exposedPortRequest := fmt.Sprintf("127.0.0.1::%s/%s", exposedPort.Port(), exposedPort.Proto())

	expectedStartLog := fmt.Sprintf("Starting Stun server, listening port at %s", stunPort)
	req := testcontainers.ContainerRequest{
		Image:        "stun:latest",
		ExposedPorts: []string{exposedPortRequest},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&logConsumer},
		},
		WaitingFor: wait.ForLog(expectedStartLog),
	}
	stunC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start app: %s", err)
	}
	defer func() {
		if err := stunC.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop app: %s", err)
		}
	}()


	// server address will be localhost:mappedPort
	serverPort, err := stunC.MappedPort(ctx, exposedPort)
	if err != nil {
		t.Error(err)
	}
	serverUri := fmt.Sprintf("%s:%s", "localhost", serverPort.Port())

	serverAddr, err := net.ResolveUDPAddr("udp", serverUri)
	if err != nil {
		log.Fatalln("failed to resolve addr: ", serverUri, "with error: ", err)
	}

	conn, err := net.Dial(serverAddr.Network(), serverAddr.String())
	if err != nil {
		log.Fatalln("failed to dial conn:", err)
	}

	log.Printf("Server :%s, %s", serverAddr.Network(), serverAddr.String())

	var options []stun.ClientOption
	if serverAddr.Network() == "tcp" {
		// Switching to "NO-RTO" mode.
		log.Println("using WithNoRetransmit for TCP") 
		options = append(options, stun.WithNoRetransmit)
	}
	client, err := stun.NewClient(conn, options...)
	if err != nil {
		log.Fatalln("failed to create client:", err)
	}

	request, err := stun.Build(stun.BindingRequest, stun.TransactionID, stun.Fingerprint)
	if err != nil {
		log.Fatalln("failed to build request:", err)
	}

    // stun server should see request coming from container gateway
	// find container gateway addr
	networks, err := stunC.Networks(ctx)
	if err != nil {
		t.Error(err)
	}
	for _, n := range networks {
		fmt.Println("networks : ", n)
	}

	dockerClient, err := dc.NewClientWithOpts()
	if nil != err {
		t.Error(err)
	}

	networkInspect, err := dockerClient.NetworkInspect(ctx, networks[0], types.NetworkInspectOptions{})
	if nil != err {
		t.Error(err)
	}
	expectedMappedIp := networkInspect.IPAM.Config[0].Gateway

	fmt.Printf("Container should see request arriving at network: %s from ip: %s", networks[0], expectedMappedIp)

	// Sending request to STUN server, waiting for response message.
	if err := client.Do(request, func(event stun.Event) {
		if event.Error != nil {
			log.Fatalln("got event with error:", event.Error)
		}

		response := event.Message
		if response.Type != stun.BindingSuccess {
			var errCode stun.ErrorCodeAttribute
			if codeErr := errCode.GetFrom(response); codeErr != nil {
				log.Fatalln("failed to get error code:", codeErr)
			}
			log.Fatalln("bad message", response, errCode)
		}

		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var xorMappedAddr stun.XORMappedAddress
		if err := response.Parse(&xorMappedAddr); err != nil {
			log.Println("failed to parse xor mapped address:", err)
		}

		// Decoding MAPPED-ADDdnetworkRESS attribute from message.
		var mappedAddr stun.MappedAddress
		if err := response.Parse(&mappedAddr); err != nil {
			log.Fatalln("failed to parse mapped address:", err)
		}

		fmt.Println("xor mapped IP is", xorMappedAddr)
		fmt.Println("mapped IP is", mappedAddr)

		if expectedMappedIp != mappedAddr.IP.String() {
			log.Fatalln(expectedMappedIp, "!=", mappedAddr)
		}

	}); err != nil {
		panic(err)
	}
}