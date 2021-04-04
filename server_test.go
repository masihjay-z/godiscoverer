package godiscoverer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestGetDefaultServer(t *testing.T) {
	defaultServer = &Server{}
	assert.Same(t, defaultServer, GetDefaultServer())
}

func TestSetDefaultServer(t *testing.T) {
	server := &Server{}
	SetDefaultServer(server)
	assert.Same(t, server, GetDefaultServer())
}

func TestNewServer(t *testing.T) {
	serviceGetterReader := &DefaultServiceGetterResponseReader{}
	serviceRegistererReader := &DefaultRegistererResponseReader{}
	serviceGetter := &DefaultServiceGetter{}
	serviceRegisterer := &DefaultServiceRegisterer{}
	server := NewServer("localhost", int64(60), serviceGetterReader, serviceRegistererReader, serviceGetter, serviceRegisterer)
	assert.Equal(t, server.Address, "localhost")
	assert.Equal(t, server.TTL, int64(60))
	assert.Equal(t,server.serviceResponseReader,serviceGetterReader)
	assert.Equal(t,server.registerResponseReader,serviceRegistererReader)
	assert.Equal(t,server.serviceGetter,serviceGetter)
	assert.Equal(t,server.serviceRegisterer,serviceRegisterer)
}

func TestServer_GetAddress(t *testing.T) {
	server:=Server{Address: "localhost"}
	assert.Equal(t,server.Address,"localhost")
}

func TestServer_SetAddress(t *testing.T) {
	server:=Server{Address: "localhost"}
	server.SetAddress("localhost-2")
	assert.Equal(t,server.Address,"localhost-2")
}

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

func TestServer_DoRegistering(t *testing.T) {
	testRegisterer := new(MockedServiceRegisterer)

	server := NewServer("localhost", 1, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, testRegisterer)
	service := NewService("test-service", "test", "80")

	testRegisterer.On("Register", &server, &service).Return(ServiceRegistererResponse{Message: "", Code: 200, IsSuccess: true, Data: int64(1)}, nil).Twice()
	ctx, _ := context.WithTimeout(context.Background(), 2100*time.Millisecond)
	server.DoRegistering(&service, ctx)
	assert.True(t, server.Registered(&service))
	testRegisterer.AssertExpectations(t)

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

func TestServer_GetRegisteredServices(t *testing.T) {
	server := NewServer("localhost", 60, &DefaultServiceGetterResponseReader{}, &DefaultRegistererResponseReader{}, &DefaultServiceGetter{}, &DefaultServiceRegisterer{})
	service := NewService("test-service", "test", "8000")
	services := map[string]int64{service.Name: int64(111)}
	server.registeredServices = services
	assert.Equal(t, server.GetRegisteredServices(), services)
}
