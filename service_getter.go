package godiscoverer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceGetter interface {
	GetServices(server *Server) (ServiceGetterResponse, error)
}

type DefaultServiceGetter struct{}

func (getter *DefaultServiceGetter) GetServices(server *Server) (ServiceGetterResponse, error) {
	res, err := http.Get(server.GetAddress())
	if err != nil {
		return ServiceGetterResponse{}, fmt.Errorf("unable to send request: %w", err)
	}
	response := newServiceResponse()
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return ServiceGetterResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}
	return response, nil
}
