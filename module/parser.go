package module

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"reflect"
)

const (
	PARSER_BASIC = iota
	PARSER_JSON
	PARSER_XML
	PARSER_YAML
	PARSER_TOML
)

type Parser interface {
	Unmarshal([]byte, any) error
	Marshal(any) ([]byte, error)
}

type parserSet map[int]Parser

func (s parserSet) isExists(key int) bool {
	_, ok := s[key]
	return ok
}

var parsers = parserSet{
	PARSER_BASIC: &ParserBasic{},
	PARSER_JSON:  &ParserJson{},
	PARSER_XML:   &ParserXml{},
	PARSER_YAML:  &ParserYaml{},
	PARSER_TOML:  &ParserToml{},
}

type ParserJson struct{}

func (p *ParserJson) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func (p *ParserJson) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

type ParserXml struct{}

func (p *ParserXml) Unmarshal(data []byte, value any) error {
	return xml.Unmarshal(data, value)
}

func (p *ParserXml) Marshal(value any) ([]byte, error) {
	return xml.Marshal(value)
}

type ParserYaml struct{}

func (p *ParserYaml) Unmarshal(data []byte, value any) error {
	return yaml.Unmarshal(data, value)
}

func (p *ParserYaml) Marshal(value any) ([]byte, error) {
	return yaml.Marshal(value)
}

type ParserToml struct{}

func (p *ParserToml) Unmarshal(data []byte, value any) error {
	return toml.Unmarshal(data, value)
}

func (p *ParserToml) Marshal(value any) ([]byte, error) {
	return toml.Marshal(value)
}

type ParserBasic struct{}

func (p *ParserBasic) Unmarshal(data []byte, value any) error {
	t := reflect.TypeOf(value)
	if t.Kind() != reflect.Pointer {
		return errors.New("value is not a pointer")
	}
	switch v := value.(type) {
	case *[]byte:
		*v = data
	case *string:
		*v = string(data)
	default:
		return errors.New("value kind is unknown")
	}
	return nil
}

func (p *ParserBasic) Marshal(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case *[]byte:
		return *v, nil
	case *string:
		return []byte(*v), nil
	default:
		return []byte(
			fmt.Sprintf("%v", value),
		), nil
	}
}
