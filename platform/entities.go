package platform

//module
type Module interface {
	ID() string
	Version() string
}

//service that can handle HTTP requests
type Service interface {
	Start() error
	Stop() error
	Setup() error
}

type ServiceModel interface {
	Create() func() error
	Migrate() func(fromVersion string) error
}
