package module

import (
	"fmt"
	"github.com/culionbear/watershed/config"
	consul "github.com/hashicorp/consul/api"
)

const (
	OPTION_CONFIG  = "config"
	OPTION_SERVICE = "service"
)

const (
	TAG_GRPC = "grpc"
	TAG_HTTP = "http"
)

type Module struct {
	conf    config.ModuleConfig
	path    string
	handler *consul.Client
}

func New(conf *config.ModuleConfig) *Module {
	return &Module{
		path:    fmt.Sprintf("%s.%s.%s.", conf.Project, conf.Name, conf.Env),
		conf:    *conf,
		handler: conf.ConsulClient,
	}
}
