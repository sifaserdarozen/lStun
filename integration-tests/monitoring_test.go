package apptest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/prometheus/common/expfmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	monitoringPort = "8081"
)

/*
	Expected Metrics will be like

...
# HELP lstun_build_info Information about lstun binary
# TYPE lstun_build_info gauge
lstun_build_info{build_date="2024-09-10T15:34:27",env="test",version="dirty-bd23b98"} 1
# HELP lstun_uptime_sec Information about binary uptime
# TYPE lstun_uptime_sec gauge
lstun_uptime_sec 6
...
*/
func TestMonitoring(t *testing.T) {
	ctx := context.Background()
	var logConsumer LogConsumer

	// docker run -p 127.0.0.1::8081/udp stun:latest
	exposedPort, err := nat.NewPort(tcpProto, monitoringPort)
	if err != nil {
		t.Error(err)
	}
	exposedPortRequest := fmt.Sprintf("127.0.0.1::%s/%s", exposedPort.Port(), exposedPort.Proto())

	expectedStartLog := fmt.Sprintf("Starting Stun server, listening port at %s/%s", stunPort, tcpProto)
	expectedEnv := "test"
	req := testcontainers.ContainerRequest{
		Image:        "stun:latest",
		ExposedPorts: []string{exposedPortRequest},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&logConsumer},
		},
		WaitingFor: wait.ForLog(expectedStartLog),
		Env:        map[string]string{"LSTN_INFO_ENV": expectedEnv},
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

	// scrape prometheus metrics
	// server address will be localhost:mappedPort
	serverPort, err := stunC.MappedPort(ctx, exposedPort)
	if err != nil {
		t.Error(err)
	}
	monitoringUri := fmt.Sprintf("%s:%s/%s", "http://localhost", serverPort.Port(), "metrics")

	resp, err := http.Get(monitoringUri)
	if err != nil {
		t.Error(err)
	}

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		t.Error(err)
	}

	uptimeSec, ok := mf["lstun_uptime_sec"]
	if !ok {
		t.Error("Metric lstun_uptim_sec is not present")
	}

	buildInfo, ok := mf["lstun_build_info"]
	if !ok {
		t.Error("Metric lstun_build_info is not present")
	}

	uptimeMetric := uptimeSec.GetMetric()[0].GetGauge().GetValue()
	if uptimeMetric < 0 {
		t.Error("Invalid uptime value")
	}

	buildLabel := buildInfo.GetMetric()[0].GetLabel()
	env := ""
	for _, v := range buildLabel {
		if v.GetName() == "env" {
			env = v.GetValue()
			break
		}
	}

	if env != expectedEnv {
		t.Errorf("env lable %s is different than expected %s", env, expectedEnv)
	}
}
