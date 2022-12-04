package module

import (
	"errors"
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

func (m *Module) GetConfig(key string) ([]byte, error) {
	rsp, _, err := m.handler.KV().Get(key, &consul.QueryOptions{
		Token: m.conf.ConsulConfig.Token,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Value, nil
}

func (m *Module) GetModuleConfig(key string) ([]byte, error) {
	key = fmt.Sprintf("%s.%s", m.path, key)
	return m.GetConfig(key)
}

func (m *Module) GetConfigWithParser(key string, parser int, value any, funcList ...func([]byte) ([]byte, error)) error {
	if !parsers.isExists(parser) {
		return errors.New("parser is not found")
	}
	data, err := m.GetConfig(key)
	if err != nil {
		return err
	}
	for _, f := range funcList {
		data, err = f(data)
		if err != nil {
			return err
		}
	}
	return parsers[parser].Unmarshal(data, value)
}

func (m *Module) GetModuleConfigWithParser(key string, parser int, value any, funcList ...func([]byte) ([]byte, error)) error {
	key = fmt.Sprintf("%s.%s", m.path, key)
	return m.GetConfigWithParser(key, parser, value, funcList...)
}

func (m *Module) SetConfig(key string, value []byte) error {
	_, err := m.handler.KV().Put(&consul.KVPair{
		Key:   key,
		Value: value,
	}, &consul.WriteOptions{
		Token: m.conf.ConsulConfig.Token,
	})
	return err
}

func (m *Module) SetModuleConfig(key string, value []byte) error {
	key = fmt.Sprintf("%s.%s", m.path, key)
	return m.SetConfig(key, value)
}

func (m *Module) SetConfigWithParser(key string, parser int, value any, funcList ...func([]byte) ([]byte, error)) error {
	if !parsers.isExists(parser) {
		return errors.New("parser is not found")
	}
	data, err := parsers[parser].Marshal(value)
	if err != nil {
		return err
	}
	for _, f := range funcList {
		data, err = f(data)
		if err != nil {
			return err
		}
	}
	return m.SetConfig(key, data)
}

func (m *Module) SetModuleConfigWithParser(key string, parser int, value any, funcList ...func([]byte) ([]byte, error)) error {
	key = fmt.Sprintf("%s.%s", m.path, key)
	return m.SetConfigWithParser(key, parser, value, funcList...)
}