package godiscoverer

type Service struct {
	Name string
	Host string
	Port string
}

func NewService(name string, host string, port string) Service {
	return Service{
		Name: name,
		Host: host,
		Port: port,
	}
}