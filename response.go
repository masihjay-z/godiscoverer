package godiscoverer

type ServiceResponse struct {
	IsSuccess bool
	Data      []Service
	Message   string
	Code      int
}

func newServiceResponse() ServiceResponse  {
	return ServiceResponse{}
}

func (response *ServiceResponse) GetServices() []Service {
	return response.Data
}

type RegisterResponse struct {
	IsSuccess bool
	Data      int64
	Message   string
	Code      int
}

func newRegisterResponse() RegisterResponse  {
	return RegisterResponse{}
}

func (response *RegisterResponse) GetTTL() int64 {
	return response.Data
}