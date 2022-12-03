package module

import (
	"fmt"
	"github.com/culionbear/watershed/config"
	consul "github.com/hashicorp/consul/api"
)

type Module struct {
	conf    config.ModuleConfig
	path    string
	handler *consul.Client
}

func New(conf *config.ModuleConfig) *Module {
	return &Module{
		path:    fmt.Sprintf("%s.%s.%s", conf.Project, conf.Env, conf.Name),
		conf:    *conf,
		handler: conf.ConsulClient,
	}
}
