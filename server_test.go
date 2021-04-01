package godiscoverer

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

func TestServer_HasServices(t *testing.T) {
	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, &DefaultServiceRegisterer{registerServiceLock: sync.Mutex{}})
	assert.False(t, server.HasServices(), "server HasServices method should return false but return true")
	server.lastGettingServices = time.Now().UnixNano()
	assert.True(t, server.HasServices(), "server HasServices method should return true but return false")
}

func TestServer_Registered(t *testing.T) {
	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, &DefaultServiceRegisterer{registerServiceLock: sync.Mutex{}})
	service := NewService("test-service", "test", "80")

	assert.False(t, server.Registered(&service), "server Registered method should return false but return true")

	server.registeredServices = map[string]int64{service.Name: time.Now().Unix()}
	assert.True(t, server.Registered(&service), "server Registered method should return true but return false, when registered list is empty.")

	server.registeredServices = map[string]int64{"wrong-service": time.Now().Unix()}
	assert.False(t, server.Registered(&service), "server Registered method should return false but return true, when registered list has a service but not asserted service.")

	server.registeredServices = map[string]int64{service.Name: time.Now().Add(-60 * time.Second).Unix()}
	assert.False(t, server.Registered(&service), "server Registered method should return false but return true, when ttl expired!")
}

type MockedServiceRegisterer struct {
	mock.Mock
}

func (registerer *MockedServiceRegisterer) Register(server *Server, service *Service) (ServiceRegistererResponse, error) {
	args := registerer.Called(server, service)
	return args.Get(0).(ServiceRegistererResponse), args.Error(1)
}

func TestServer_ForceRegister(t *testing.T) {
	testRegisterer := new(MockedServiceRegisterer)

	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, testRegisterer)
	service := NewService("test-service", "test", "80")

	testRegisterer.On("Register", &server, &service).Return(ServiceRegistererResponse{Message: "", Code: 200, IsSuccess: true, Data: int64(300)}, nil).Once()
	server.ForceRegister(&service)
	assert.True(t, server.Registered(&service))
	assert.Equal(t, server.TTL, int64(300))
	testRegisterer.AssertExpectations(t)

	server.registeredServices = map[string]int64{}
	testRegisterer.On("Register", &server, &service).Return(ServiceRegistererResponse{Message: "", Code: 400, IsSuccess: false, Data: nil}, nil)
	server.ForceRegister(&service)
	assert.False(t, server.Registered(&service))
	testRegisterer.AssertExpectations(t)
}

func TestServer_Register(t *testing.T) {
	testRegisterer := new(MockedServiceRegisterer)

	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, testRegisterer)
	service := NewService("test-service", "test", "80")

	testRegisterer.On("Register", &server, &service).Return(ServiceRegistererResponse{Message: "", Code: 200, IsSuccess: true, Data: int64(300)}, nil).Once()
	server.Register(&service)
	assert.True(t, server.Registered(&service))
	testRegisterer.AssertExpectations(t)

	server.Register(&service)
	assert.True(t, server.Registered(&service))
}

type MockedServiceGetter struct {
	mock.Mock
}

func (getter *MockedServiceGetter) GetServices(server *Server) (ServiceGetterResponse, error) {
	args := getter.Called(server)
	return args.Get(0).(ServiceGetterResponse), args.Error(1)
}

func TestServer_Find(t *testing.T) {
	testGetter := new(MockedServiceGetter)

	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, testGetter, &DefaultServiceRegisterer{})
	service := NewService("test-service", "test", "8000")

	testGetter.On("GetServices", &server).Return(ServiceGetterResponse{Message: "", Code: 200, IsSuccess: true, Data: []Service{{Name: service.Name, Port: service.Port, Host: service.Host}}}, nil).Once()
	gottenService, _ := server.Find("test-service")
	assert.Equal(t, gottenService.Port, service.Port)
	testGetter.AssertExpectations(t)


	gottenServiceFromCache, _ := server.Find("test-service")
	assert.Equal(t, gottenServiceFromCache.Port, service.Port)

	server.TTL = 0
	testGetter.On("GetServices", &server).Return(ServiceGetterResponse{Message: "", Code: 200, IsSuccess: true, Data: []Service{{Name: service.Name, Port: service.Port, Host: service.Host}}}, nil).Once()
	gottenServiceWhenTTLExpired, _ := server.Find("test-service")
	assert.Equal(t, gottenServiceWhenTTLExpired.Port, service.Port)
	testGetter.AssertExpectations(t)
}
