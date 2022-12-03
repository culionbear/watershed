package http

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	consul "github.com/hashicorp/consul/api"
)

type Request struct {
	Scheme string // ["http","https"]
	Method string // http.MethodXXX
	Path   string // service path
	Option func(*resty.Request)
}

type ClientConn struct {
	serviceName   string
	centerHandler *consul.Client
	mu            sync.RWMutex
	services      []string
	pointer       int
	conf          *consul.Config
	httpHandler   *resty.Client
}

func New(serviceName string, handler *consul.Client, conf *consul.Config) *ClientConn {
	m := &ClientConn{
		serviceName:   serviceName,
		centerHandler: handler,
		pointer:       0,
		conf:          conf,
		httpHandler:   resty.New(),
	}
	go m.balance()
	return m
}

func (m *ClientConn) balance() {
	for {
		list := m.getServiceAddressList()
		m.mu.Lock()
		m.services = list
		m.mu.Unlock()
		time.Sleep(time.Second * 15)
	}
}

func (m *ClientConn) getServiceAddressList() []string {
	rsp, _, err := m.centerHandler.Catalog().Service(
		m.serviceName,
		"",
		&consul.QueryOptions{
			Token: m.conf.Token,
		},
	)
	if err != nil {
		return []string{}
	}
	list := make([]string, 0)
	for _, v := range rsp {
		if v.Checks.AggregatedStatus() == consul.HealthPassing {
			list = append(list, fmt.Sprintf("%s:%d", v.ServiceAddress, v.ServicePort))
		}
	}
	return list
}

func (m *ClientConn) GetAddress() (string, error) {
	defer func() {
		m.mu.RUnlock()

		m.mu.Lock()
		if length := len(m.services); length != 0 {
			m.pointer = (m.pointer + 1) % length
		}
		m.mu.Unlock()
	}()

	m.mu.RLock()
	if m.services == nil || len(m.services) == 0 {
		return "", errors.New("address list is empty")
	}
	return m.services[m.pointer%len(m.services)], nil
}

func (m *ClientConn) Execute(request *Request) (*resty.Response, error) {
	if request.Method == "" {
		request.Method = http.MethodGet
	}
	if request.Path == "" || request.Path[0] != '/' {
		request.Path = "/" + request.Path
	}
	if request.Scheme == "" {
		request.Scheme = "http"
	}
	r := m.httpHandler.R()
	if request.Option != nil {
		request.Option(r)
	}
	
	address, err := m.GetAddress()
	if err != nil {
		return nil, err
	}
	return r.Execute(request.Method, fmt.Sprintf("%s//%s%s", request.Scheme, address, request.Path))
}
