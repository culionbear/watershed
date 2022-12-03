package module

import (
	"fmt"

	"github.com/culionbear/watershed/http"
	"google.golang.org/grpc"
	_ "github.com/mbobakov/grpc-consul-resolver"
)

func (m *Module) NewGrpcClientConn() (*grpc.ClientConn, error) {
	url := fmt.Sprintf("consul://%s/%s?healthy=true", m.conf.WatershedConfig.ConsulConfig.Address, m.path)
	if m.conf.WatershedConfig.ConsulConfig.Token != "" {
		url += fmt.Sprintf("&token=%s", m.conf.WatershedConfig.ConsulConfig.Token)
	}
	return grpc.Dial(
		url,
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
}

func (m *Module) NewHttpClientConn() *http.ClientConn {
	return http.New(m.path, m.handler, m.conf.ConsulConfig)
}
