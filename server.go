package godiscoverer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Server struct {
	Address             string
	TTL                 int64
	Services            []Service
	maxRegisteringRetry int64
	lastGettingServices int64
	registeredServices  map[string]int64
}

var defaultServer *Server

func SetDefaultServer(server *Server) {
	defaultServer = server
}

func GetDefaultServer() *Server {
	return defaultServer
}

func NewServer(address string, ttl int64) Server {
	return Server{Address: address, TTL: ttl, registeredServices: make(map[string]int64)}
}

func (server *Server) GetAddress() string {
	return server.Address
}

func (server *Server) SetAddress(address string) *Server {
	server.Address = address
	return server
}

func (server *Server) GetServices() ([]Service, error) {
	if server.HasServices() {
		return server.Services, nil
	}
	return server.ForceGetServices()
}

func (server *Server) ForceGetServices() ([]Service, error) {
	response, err := server.servicesRequest()
	if err != nil {
		return nil, fmt.Errorf("unable to getting services: %w", err)
	}
	server.Services = response.GetServices()
	server.lastGettingServices = time.Now().Unix()
	return server.Services, nil
}

func (server *Server) servicesRequest() (ServiceResponse, error) {
	res, err := http.Get(server.GetAddress())
	if err != nil {
		return ServiceResponse{}, fmt.Errorf("unable to send request: %w", err)
	}
	response := newServiceResponse()
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return ServiceResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}
	return response, nil
}

func (server *Server) HasServices() bool {
	return time.Now().Unix() < server.lastGettingServices+server.TTL
}

func (server *Server) Register(service *Service) (bool, error) {
	if server.Registered(service) {
		return true, nil
	}
	return server.ForceRegister(service)
}

func (server *Server) ForceRegister(service *Service) (bool, error) {
	response, err := server.registerRequest(service)
	if err != nil {
		return false, fmt.Errorf("unable to register: %w", err)
	}
	if response.IsSuccess {
		server.TTL = response.GetTTL()
		server.updateRegisteredServices(service)
		return true, nil
	}
	return false, nil
}

func (server *Server) DoRegistering(service *Service, ctx context.Context) {
	for {
		select {
		case <-time.Tick(time.Duration(server.TTL) * time.Second):
			_, err := server.Register(service)
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (server *Server) Registered(service *Service) bool {
	for serviceName, registrationTime := range server.GetRegisteredServices() {
		if service.Name == serviceName && time.Now().Unix() < registrationTime {
			return true
		}
	}
	return false
}

func (server *Server) registerRequest(service *Service) (RegisterResponse, error) {
	data := url.Values{}
	data.Set("name", service.Name)
	data.Set("host", service.Host)
	data.Set("port", service.Port)
	res, err := http.PostForm(server.GetAddress(), data)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("unable to send request: %w", err)
	}
	response := newRegisterResponse()
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("unable to parse json: %w", err)
	}
	return response, nil
}

func (server *Server) Find(name string) (Service, error) {
	services, err := server.GetServices()
	if err != nil {
		return Service{}, fmt.Errorf("unable to find %v service: %w", name, err)
	}
	for i := range services {
		if services[i].Name == name {
			return services[i], nil
		}
	}
	return Service{}, errors.New("service not found")
}

func (server *Server) updateRegisteredServices(service *Service) {
	server.registeredServices[service.Name] = time.Now().Unix()
}

func (server *Server) GetRegisteredServices() map[string]int64 {
	return server.registeredServices
}
