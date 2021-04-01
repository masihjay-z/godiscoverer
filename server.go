package godiscoverer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Server struct {
	Address                string
	TTL                    int64
	Services               []Service
	serviceResponseReader  ServiceGetterResponseReader
	registerResponseReader RegistererResponseReader
	serviceGetter          ServiceGetter
	serviceRegisterer      ServiceRegisterer
	lastGettingServices    int64
	registeredServices     map[string]int64
	updateServiceLock      sync.Mutex
}

var defaultServer *Server

func SetDefaultServer(server *Server) {
	defaultServer = server
}

func GetDefaultServer() *Server {
	return defaultServer
}

func NewServer(address string, ttl int64, serviceResponseReader ServiceGetterResponseReader, registerResponseReader RegistererResponseReader, serviceGetter ServiceGetter, serviceRegisterer ServiceRegisterer) Server {
	return Server{
		Address:                address,
		TTL:                    ttl,
		serviceResponseReader:  serviceResponseReader,
		registerResponseReader: registerResponseReader,
		serviceGetter:          serviceGetter,
		serviceRegisterer:      serviceRegisterer,
		registeredServices:     make(map[string]int64),
		updateServiceLock:      sync.Mutex{},
	}
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
	response, err := server.serviceGetter.GetServices(server)
	if err != nil {
		return nil, fmt.Errorf("unable to getting services: %w", err)
	}
	server.Services = server.serviceResponseReader.GetServices(&response)
	server.lastGettingServices = time.Now().Unix()
	return server.Services, nil
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
	response, err := server.serviceRegisterer.Register(server, service)
	if err != nil {
		return false, fmt.Errorf("unable to register: %w", err)
	}
	if response.IsSuccess {
		server.TTL = server.registerResponseReader.GetTTL(&response)
		server.updateRegisteredServices(service)
		log.Printf("%v service register successfully", service.Name)
		return true, nil
	}
	return false, nil
}

func (server *Server) DoRegistering(service *Service, ctx context.Context) {
	for {
		updateCtx, _ := context.WithTimeout(context.Background(), time.Duration(server.TTL)*time.Second)
		select {
		case <-updateCtx.Done():
			res, err := server.Register(service)
			if err != nil || res == false {
				log.Printf("failed to register %v:%v\n", service.Name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (server *Server) Registered(service *Service) bool {
	for serviceName, registrationTime := range server.GetRegisteredServices() {
		if service.Name == serviceName && time.Now().Unix() < registrationTime+server.TTL {
			return true
		}
	}
	return false
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
	server.updateServiceLock.Lock()
	server.registeredServices[service.Name] = time.Now().Unix()
	server.updateServiceLock.Unlock()
}

func (server *Server) GetRegisteredServices() map[string]int64 {
	return server.registeredServices
}
