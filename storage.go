package registry_frontend

type Storage interface {
	Registries() ([]Registry, error)
	Registry(name string) (Registry, error)
	Add(r Registry) error
	Update(r Registry) error
	Clear() error // ???
	Remove(r Registry) error
}
