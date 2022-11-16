package watershed

import (
	"errors"
	"github.com/culionbear/watershed/config"
	"github.com/culionbear/watershed/module"
	consul "github.com/hashicorp/consul/api"
)

type Manager struct {
	conf    config.WatershedConfig
	handler *consul.Client
}

func New(conf *config.WatershedConfig) (*Manager, error) {
	if conf.Project == "" {
		return nil, errors.New("project name is empty")
	}
	handler, err := consul.NewClient(conf.ConsulConfig)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		conf:    *conf,
		handler: handler,
	}
	if m.conf.Env == "" {
		m.conf.Env = "production"
	}
	return m, nil
}

func (m *Manager) NewModule(moduleName string) *module.Module {
	return module.New(&config.ModuleConfig{
		WatershedConfig: m.conf,
		Name:            moduleName,
		ConsulClient:    m.handler,
	})
}

func (m *Manager) GetConsulClient() *consul.Client {
	return m.handler
}
