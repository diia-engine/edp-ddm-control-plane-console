package gitserver

type ServiceInterface interface {
	Get(name string) (*GitServer, error)
}
