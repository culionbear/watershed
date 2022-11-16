package config

import consul "github.com/hashicorp/consul/api"

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
