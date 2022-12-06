package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/culionbear/watershed/config"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	SERVICE_TYPE_GRPC = "grpc"
	SERVICE_TYPE_HTTP = "http"
)

const (
	AUTO_IP_DONOTHING = iota
	AUTO_IP
	AUTO_IP_PUBLIC
)

var autoIpDefaultStore = map[int]func() (net.IP, error){
	AUTO_IP_DONOTHING: nil,
	AUTO_IP:           getPrivateIp,
	AUTO_IP_PUBLIC:    getPublicIp,
}

func getPublicIp() (net.IP, error) {
	rsp, err := http.Get("https://www.ip.cn/api/index?ip=&type=0")
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	msg, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	var body struct {
		Rs       int    `json:"rs"`
		Code     int    `json:"code"`
		Address  string `json:"address"`
		Ip       string `json:"ip"`
		IsDomain int    `json:"isDomain"`
	}
	err = json.Unmarshal(msg, &body)
	ip := net.ParseIP(body.Ip)
	if ip == nil {
		return nil, errors.New("ip is error from ip.cn")
	}
	return ip, nil
}

func getPrivateIp() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

func compareIpInLocalIpList(ip net.IP) bool {
	if ip == nil {
		return false
	}
	ipList, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, addr := range ipList {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip.Equal(ipNet.IP) {
				return true
			}
		}
	}
	return false
}

func (m *Module) initIp(conf *config.ServiceConfig) (net.IP, error) {
	if _, ok := autoIpDefaultStore[conf.AutoGetIPModel]; !ok {
		return nil, errors.New("auto get ip module type is unknown")
	}
	defaultIp := net.IP{0, 0, 0, 0}
	if conf.AutoGetIPModel == AUTO_IP_DONOTHING {
		if conf.IP == nil || conf.IP.Equal(defaultIp) {
			return nil, errors.New("ip is default")
		}
		if (!conf.IsListenerIpDefault) && compareIpInLocalIpList(conf.IP) {
			return conf.IP, nil
		}
		return defaultIp, nil
	}
	var err error
	if conf.CustomIPConsumer != nil {
		conf.IP, err = conf.CustomIPConsumer()
		if err != nil {
			return nil, err
		}
		if conf.IP == nil {
			return nil, errors.New("ip is default")
		}
	} else {
		conf.IP, err = autoIpDefaultStore[conf.AutoGetIPModel]()
		if err != nil {
			return nil, err
		}
	}
	if (!conf.IsListenerIpDefault) && compareIpInLocalIpList(conf.IP) {
		return conf.IP, nil
	}
	return defaultIp, nil
}

func (m *Module) initServiceRegisterConfig(conf *config.ServiceConfig, serviceType string) (net.IP, error) {
	if conf == nil {
		return nil, errors.New("service config is nil")
	}
	serviceAddress, err := m.initIp(conf)
	if err != nil {
		return nil, err
	}
	if conf.Port <= 0 {
		conf.Port = 80
	}
	if conf.RegisterConfig == nil {
		conf.RegisterConfig = new(consul.AgentServiceRegistration)
	}
	conf.RegisterConfig.ID = fmt.Sprintf("%s.%s", m.path, uuid.New().String())
	conf.RegisterConfig.Name = m.path
	if conf.RegisterConfig.Tags == nil {
		conf.RegisterConfig.Tags = make([]string, 0)
	}
	conf.RegisterConfig.Tags = append(conf.RegisterConfig.Tags, serviceType)
	conf.RegisterConfig.Port = conf.Port
	conf.RegisterConfig.Address = conf.IP.String()
	if conf.RegisterConfig.Check == nil {
		conf.RegisterConfig.Check = new(consul.AgentServiceCheck)
	}
	if conf.RegisterConfig.Check.Interval == "" {
		conf.RegisterConfig.Check.Interval = "5s"
	}
	if conf.RegisterConfig.Check.DeregisterCriticalServiceAfter == "" {
		conf.RegisterConfig.Check.DeregisterCriticalServiceAfter = "10s"
	}
	url := fmt.Sprintf("%s:%d/health", conf.IP.String(), conf.Port)
	switch serviceType {
	case SERVICE_TYPE_GRPC:
		conf.RegisterConfig.Check.GRPC = url
	case SERVICE_TYPE_HTTP:
		conf.RegisterConfig.Check.HTTP = url
	}
	return serviceAddress, nil
}

func (m *Module) GrpcService(conf *config.GrpServiceConfig) error {
	serviceAddress, err := m.initServiceRegisterConfig(&conf.ServiceConfig, SERVICE_TYPE_GRPC)
	if err != nil {
		return err
	}
	err = m.handler.Agent().ServiceRegister(conf.RegisterConfig)
	if err != nil {
		return err
	}
	service := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(service, new(HealthImpl))
	if conf.CallBackFunc != nil {
		err = conf.CallBackFunc(service)
		if err != nil {
			return err
		}
	}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   serviceAddress,
		Port: conf.Port,
	})
	return service.Serve(listener)
}

func (m *Module) HttpService(conf *config.HttpServiceConfig) error {
	if conf.Runner == nil {
		return errors.New("runner is nil")
	}
	serviceAddress, err := m.initServiceRegisterConfig(&conf.ServiceConfig, SERVICE_TYPE_HTTP)
	if err != nil {
		return err
	}
	err = m.handler.Agent().ServiceRegister(conf.RegisterConfig)
	if err != nil {
		return err
	}
	return conf.Runner(serviceAddress, conf.Port, "/health", m.httpHealthCheck)
}

func (m *Module) httpHealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"msg":"success"}`))
}
