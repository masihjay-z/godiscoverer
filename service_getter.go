package godiscoverer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceGetter interface {
	GetServices(server *Server) (ServiceResponse, error)
}

type DefaultServiceGetter struct{}

func (getter *DefaultServiceGetter) GetServices(server *Server) (ServiceResponse, error) {
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
