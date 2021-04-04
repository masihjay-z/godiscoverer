package godiscoverer

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/url"
	"sync"
)

type ServiceRegisterer interface {
	Register(server *Server, service *Service) (ServiceRegistererResponse, error)
}

type DefaultServiceRegisterer struct{
	registerServiceLock    sync.Mutex
}

func (registerer *DefaultServiceRegisterer) Register(server *Server, service *Service) (ServiceRegistererResponse, error) {
	data := url.Values{}
	data.Set("name", service.Name)
	data.Set("host", service.Host)
	data.Set("port", service.Port)
	registerer.registerServiceLock.Lock()
	res, err := http.PostForm(server.GetAddress(), data)
	registerer.registerServiceLock.Unlock()
	if err != nil {
		return ServiceRegistererResponse{}, fmt.Errorf("unable to send request: %w", err)
	}
	response := newRegisterResponse()
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return ServiceRegistererResponse{}, fmt.Errorf("unable to parse json: %w", err)
	}
	return response, nil
}

type MockedServiceRegisterer struct {
	mock.Mock
}

func (registerer *MockedServiceRegisterer) Register(server *Server, service *Service) (ServiceRegistererResponse, error) {
	args := registerer.Called(server, service)
	return args.Get(0).(ServiceRegistererResponse), args.Error(1)
}
