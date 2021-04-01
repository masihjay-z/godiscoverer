package godiscoverer

type RegisterResponse struct {
	IsSuccess bool
	Data      interface{}
	Message   string
	Code      int
}

func newRegisterResponse() RegisterResponse {
	return RegisterResponse{}
}

type RegisterResponseReader interface {
	GetTTL(registerResponse *RegisterResponse) int64
}

type DefaultRegisterResponseReader struct{}

func (response *DefaultRegisterResponseReader) GetTTL(registerResponse *RegisterResponse) int64 {
	return registerResponse.Data.(int64)
}
