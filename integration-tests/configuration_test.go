package apptest

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	dc "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pion/stun/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConfiguration(t *testing.T) {

	// check test config
	testConfigPath, err := filepath.Abs(filepath.Join(".", "config", "stun.yaml"))
	if err != nil {
		t.Errorf("Test file path can not be generated with error: %s", err)
	}

	testCases := map[string]struct {
		port string
		file []testcontainers.ContainerFile
		env  map[string]string
		cmd  []string
	}{
		"default should provide udp service at standard stun port 3478": {
			port: stunPort,
			env:  map[string]string{},
			file: []testcontainers.ContainerFile{},
			cmd:  []string{"./stun"},
		},
		"configuration file should override defaults and provide udp service at 4444": {
			port: "4444",
			env:  map[string]string{},
			file: []testcontainers.ContainerFile{{HostFilePath: testConfigPath, ContainerFilePath: "/stun/config/stun.yaml"}},
			cmd:  []string{"./stun"},
		},
		"environments should override configuration file reads and provide udp service at 5555": {
			port: "5555",
			env:  map[string]string{"LSTN_UDP_PORT": "5555"},
			file: []testcontainers.ContainerFile{{HostFilePath: testConfigPath, ContainerFilePath: "/stun/config/stun.yaml"}},
			cmd:  []string{"./stun"},
		},
		"cli flags should override all and provide udp service at 6666": {
			port: "6666",
			env:  map[string]string{"LSTN_UDP_PORT": "5555"},
			file: []testcontainers.ContainerFile{{HostFilePath: testConfigPath, ContainerFilePath: "/stun/config/stun.yaml"}},
			cmd:  []string{"./stun", "--udp-port", "6666"},
		},
	}

	for name, test := range testCases {
		// test := test // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var logConsumer LogConsumer

			// docker run -p 127.0.0.1::3478/udp stun:latest
			exposedPort, err := nat.NewPort(udpProto, test.port)
			if err != nil {
				t.Error(err)
			}
			exposedPortRequest := fmt.Sprintf("127.0.0.1::%s/%s", exposedPort.Port(), exposedPort.Proto())

			expectedStartLog := fmt.Sprintf("Starting Stun server, listening port at %s/%s", test.port, udpProto)
			req := testcontainers.ContainerRequest{
				Image:        "stun:latest",
				ExposedPorts: []string{exposedPortRequest},
				LogConsumerCfg: &testcontainers.LogConsumerConfig{
					Consumers: []testcontainers.LogConsumer{&logConsumer},
				},
				WaitingFor: wait.ForLog(expectedStartLog),
				Env:        test.env,
				Files:      test.file,
				Cmd:        test.cmd,
			}
			stunC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err != nil {
				t.Errorf("Could not start app: %s", err)
			}
			defer func() {
				if err := stunC.Terminate(ctx); err != nil {
					t.Errorf("Could not stop app: %s", err)
				}
			}()

			// server address will be localhost:mappedPort
			serverPort, err := stunC.MappedPort(ctx, exposedPort)
			if err != nil {
				t.Error(err)
			}
			serverUri := fmt.Sprintf("%s:%s", "localhost", serverPort.Port())

			serverNet, severAdr, err := udpResolver(udpProto, serverUri)
			if err != nil {
				t.Errorf("failed to resolve addr: %s with error: %s", serverUri, err)
			}

			conn, err := net.Dial(serverNet, severAdr)
			if err != nil {
				t.Errorf("failed to dial conn: %s", err)
			}

			t.Logf("Server :%s, %s", serverNet, severAdr)

			var options []stun.ClientOption
			if serverNet == "tcp" {
				// Switching to "NO-RTO" mode.
				t.Log("using WithNoRetransmit for TCP")
				options = append(options, stun.WithNoRetransmit)
			}
			client, err := stun.NewClient(conn, options...)
			if err != nil {
				t.Errorf("failed to create client: %s", err)
			}

			request, err := stun.Build(stun.BindingRequest, stun.TransactionID, stun.Fingerprint)
			if err != nil {
				t.Errorf("failed to build request: %s", err)
			}

			// stun server should see request coming from container gateway
			// find container gateway addr
			networks, err := stunC.Networks(ctx)
			if err != nil {
				t.Error(err)
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

			t.Logf("Container should see request arriving at network: %s from ip: %s\n", networks[0], expectedMappedIp)

			// Sending request to STUN server, waiting for response message.
			if err := client.Do(request, func(event stun.Event) {
				if event.Error != nil {
					t.Errorf("Got event with error: %s", event.Error)
				}

				response := event.Message
				if response.Type != stun.BindingSuccess {
					var errCode stun.ErrorCodeAttribute
					if codeErr := errCode.GetFrom(response); codeErr != nil {
						t.Errorf("failed to get error code: %s", codeErr)
					}
					t.Errorf("Unexpected response %s, with code %s", response, errCode)
				}

				// Decoding XOR-MAPPED-ADDRESS attribute from message.
				var xorMappedAddr stun.XORMappedAddress
				if err := response.Parse(&xorMappedAddr); err != nil {
					t.Logf("Failed to parse xor mapped address (OPTIONAL): %s", err)
				}

				// Decoding MAPPED-ADDdnetworkRESS attribute from message.
				var mappedAddr stun.MappedAddress
				if err := response.Parse(&mappedAddr); err != nil {
					t.Errorf("Failed to parse mapped address: %s", err)
				}

				t.Logf("xor mapped IP is %s", xorMappedAddr)
				t.Logf("mapped IP is %s", mappedAddr)

				if expectedMappedIp != mappedAddr.IP.String() {
					t.Errorf("expected ip = %s != %s = mapped ip", expectedMappedIp, mappedAddr)
				}
			}); err != nil {
				t.Errorf("Error in stun request %s", err)
			}
		})
	}
}
