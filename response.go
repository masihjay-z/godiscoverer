package godiscoverer

type ServiceResponse struct {
	IsSuccess bool
	Data      []Service
	Message   string
	Code      int
}

func (response *ServiceResponse) GetServices() []Service {
	return response.Data
}

func NewServiceResponse() ServiceResponse  {
	return ServiceResponse{}
}

type RegisterResponse struct {
	IsSuccess bool
	Data      int64
	Message   string
	Code      int
}

func (response *RegisterResponse) GetTTL() int64 {
	return response.Data
}

func NewRegisterResponse() RegisterResponse  {
	return RegisterResponse{}
}