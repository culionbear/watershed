package module

import (
	"errors"
	consul "github.com/hashicorp/consul/api"
)

func (m *Module) GetConfig(key string) ([]byte, error) {
	rsp, _, err := m.handler.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}
	return rsp.Value, nil
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

func (m *Module) SetConfig(key string, value []byte) error {
	_, err := m.handler.KV().Put(&consul.KVPair{
		Key:   key,
		Value: value,
	}, nil)
	return err
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
	_, err = m.handler.KV().Put(&consul.KVPair{
		Key:   key,
		Value: data,
	}, nil)
	return err
}
