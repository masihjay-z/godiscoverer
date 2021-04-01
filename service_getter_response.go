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
	return serviceGetterResponse.Data.([]Service)
}
