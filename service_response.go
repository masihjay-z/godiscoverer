package godiscoverer


type ServiceResponse struct {
	IsSuccess bool
	Data      interface{}
	Message   string
	Code      int
}

func newServiceResponse() ServiceResponse {
	return ServiceResponse{}
}

type ServiceResponseReader interface {
	GetServices(serviceResponse *ServiceResponse) []Service
}

type DefaultServiceResponseReader struct {}

func (response *DefaultServiceResponseReader) GetServices(serviceResponse *ServiceResponse) []Service {
	return serviceResponse.Data.([]Service)
}