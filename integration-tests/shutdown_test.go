package apptest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	WAIT_TIME_SEC = 1
)

func TestGracefulShutdown(t *testing.T) {
	ctx := context.Background()
	var logConsumer LogConsumer

	// docker run -p 127.0.0.1::3478/udp stun:latest
	exposedPort, err := nat.NewPort(tcpProto, stunPort)
	if err != nil {
		t.Error(err)
	}
	exposedPortRequest := fmt.Sprintf("127.0.0.1::%s/%s", exposedPort.Port(), exposedPort.Proto())

	expectedStartLog := fmt.Sprintf("Starting Stun server, listening port at %s/%s", stunPort, tcpProto)
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
		t.Errorf("Could not start app: %s", err)
	}
	defer func() {
		if err := stunC.Terminate(ctx); err != nil {
			t.Errorf("Could not stop app: %s", err)
		}
	}()

	// request graceful termination
	waitTime := WAIT_TIME_SEC * time.Second
	err = stunC.Stop(ctx, &waitTime)
	if err != nil {
		t.Errorf("Could not gracefully stop app: %s in %d sec", err, waitTime)
	}
}
