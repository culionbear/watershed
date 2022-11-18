package config

import (
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"net"
)

type WatershedConfig struct {
	Project      string
	Env          string
	ConsulConfig *consul.Config
}

type ModuleConfig struct {
	WatershedConfig
	Name         string
	ConsulClient *consul.Client
}

type ServiceConfig struct {
	IP                  net.IP
	Port                int
	IsListenerIpDefault bool
	AutoGetIPModel      int
	CustomIPConsumer    func() (net.IP, error)
	RegisterConfig      *consul.AgentServiceRegistration
}

type GrpServiceConfig struct {
	ServiceConfig
	CallBackFunc func(*grpc.Server) error
}
