package godiscoverer

type ServiceGetterResponse struct {
	IsSuccess bool
	Data      interface{}
	Message   string
	Code      int
}

func newServiceResponse() ServiceGetterResponse {
	return ServiceGetterResponse{}
}

type ServiceGetterResponseReader interface {
	GetServices(serviceGetterResponse *ServiceGetterResponse) []Service
}

type DefaultServiceGetterResponseReader struct{}

func (response *DefaultServiceGetterResponseReader) GetServices(serviceGetterResponse *ServiceGetterResponse) []Service {
	var services []Service
	interfaces := serviceGetterResponse.Data.([]interface{})
	for i := range interfaces {
		serviceInterface := interfaces[i].(map[string]interface{})
		services = append(services, Service{Host: serviceInterface["host"].(string), Port: serviceInterface["port"].(string), Name: serviceInterface["name"].(string)})
	}
	return services
}
