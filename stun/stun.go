package stun

import (
	"context"
	"sync"
)

func Start(conf *Configuration, ctx context.Context, wg *sync.WaitGroup) {
	InitInfo()
	MonitoringStart(ctx, conf.Monitoring, wg)
	UdpStart(ctx, conf.Udp, wg)
	TcpStart(ctx, conf.Tcp, wg)
}
