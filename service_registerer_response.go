package godiscoverer

type ServiceRegistererResponse struct {
	IsSuccess bool
	Data      interface{}
	Message   string
	Code      int
}

func newRegisterResponse() ServiceRegistererResponse {
	return ServiceRegistererResponse{}
}

type RegistererResponseReader interface {
	GetTTL(registererResponse *ServiceRegistererResponse) int64
}

type DefaultRegistererResponseReader struct{}

func (response *DefaultRegistererResponseReader) GetTTL(registererResponse *ServiceRegistererResponse) int64 {
	return registererResponse.Data.(int64)
}
