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
	interfaces := serviceGetterResponse.Data.([]map[string]interface{})
	for i := range interfaces {
		services = append(services, Service{Host: interfaces[i]["host"].(string), Port: interfaces[i]["port"].(string), Name: interfaces[i]["name"].(string)})
	}
	return services
}
